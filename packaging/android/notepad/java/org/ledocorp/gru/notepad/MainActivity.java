package org.ledocorp.gru.notepad;

import android.app.NativeActivity;
import android.content.Context;
import android.content.Intent;
import android.database.Cursor;
import android.graphics.Color;
import android.graphics.Insets;
import android.graphics.Rect;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.VibrationEffect;
import android.os.Vibrator;
import android.os.VibratorManager;
import android.provider.OpenableColumns;
import android.text.Editable;
import android.text.InputType;
import android.text.TextWatcher;
import android.view.KeyEvent;
import android.view.View;
import android.view.ViewGroup;
import android.view.Window;
import android.view.WindowInsets;
import android.view.WindowInsetsAnimation;
import android.view.WindowManager;
import android.view.inputmethod.EditorInfo;
import android.view.inputmethod.InputMethodManager;
import android.widget.EditText;
import android.widget.FrameLayout;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.lang.reflect.Method;
import java.nio.charset.StandardCharsets;
import java.util.List;

/**
 * NativeActivity host: system-bar tint, soft-IME proxy (chars → JNI), light haptics,
 * SAF open/save pickers.
 */
public class MainActivity extends NativeActivity {
    private static final int GRU_BG = Color.rgb(228, 229, 235);
    private static final int REQ_OPEN_DOCUMENT = 0x47524f31; // 'GRO1'
    private static final int REQ_CREATE_DOCUMENT = 0x47524f32; // 'GRO2'

    static {
        // NativeActivity also loads via android.app.lib_name; explicit load ensures
        // JNI_OnLoad RegisterNatives runs before any IME proxy callbacks.
        try {
            System.loadLibrary("grunotepad");
        } catch (UnsatisfiedLinkError ignored) {
        }
    }

    private EditText imeProxy;
    private boolean imeSuppress;
    private boolean imeWanted;
    private int lastImeHeight;
    private byte[] pendingSaveBytes;
    private String pendingSaveName;

    public native void nativeImeCodepoint(int codepoint);

    public native void nativeImeKey(int androidKeyCode);

    public native void nativeImeHeight(int px);

    public native void nativeFileOpenResult(String path, int cancelled);

    public native void nativeFileSaveResult(String path, int cancelled);

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Keep NativeActivity GLES surface size stable when the IME opens.
        getWindow()
                .setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_ADJUST_NOTHING);
        applyLightSystemBars();
        ensureImeProxy();
        installImeHeightListener();
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            applyLightSystemBars();
            if (imeWanted) {
                showGruIme();
            }
        }
    }

    /** Called from Go via JNI when a text field gains focus. */
    public void showGruIme() {
        imeWanted = true;
        runOnUiThread(
                () -> {
                    ensureImeProxy();
                    imeProxy.setFocusable(true);
                    imeProxy.setFocusableInTouchMode(true);
                    imeProxy.requestFocus();
                    InputMethodManager imm =
                            (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
                    if (imm != null) {
                        imm.showSoftInput(imeProxy, InputMethodManager.SHOW_IMPLICIT);
                    }
                    imeProxy.postDelayed(this::pollImeHeight, 120);
                    imeProxy.postDelayed(this::pollImeHeight, 360);
                });
    }

    /** Called from Go via JNI when text focus is cleared. */
    public void hideGruIme() {
        imeWanted = false;
        runOnUiThread(
                () -> {
                    if (imeProxy == null) {
                        reportImeHeight(0);
                        return;
                    }
                    InputMethodManager imm =
                            (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
                    if (imm != null) {
                        imm.hideSoftInputFromWindow(imeProxy.getWindowToken(), 0);
                    }
                    imeProxy.clearFocus();
                    reportImeHeight(0);
                });
    }

    /** Called from Go OpenBrowser — launch system browser / Custom Tabs handler. */
    public void openGruUrl(String url) {
        if (url == null || url.isEmpty()) {
            return;
        }
        runOnUiThread(
                () -> {
                    try {
                        Intent i = new Intent(Intent.ACTION_VIEW, Uri.parse(url));
                        i.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
                        startActivity(i);
                    } catch (Exception ignored) {
                    }
                });
    }

    /** Short confirmation tick (bottom-nav change). */
    public void hapticTick() {
        runOnUiThread(
                () -> {
                    try {
                        if (Build.VERSION.SDK_INT >= 31) {
                            VibratorManager vm =
                                    (VibratorManager) getSystemService(Context.VIBRATOR_MANAGER_SERVICE);
                            if (vm != null) {
                                vm.getDefaultVibrator()
                                        .vibrate(
                                                VibrationEffect.createOneShot(
                                                        18, VibrationEffect.DEFAULT_AMPLITUDE));
                            }
                        } else {
                            Vibrator v = (Vibrator) getSystemService(Context.VIBRATOR_SERVICE);
                            if (v != null && Build.VERSION.SDK_INT >= 26) {
                                v.vibrate(
                                        VibrationEffect.createOneShot(
                                                18, VibrationEffect.DEFAULT_AMPLITUDE));
                            } else if (v != null) {
                                v.vibrate(18);
                            }
                        }
                    } catch (Exception ignored) {
                    }
                });
    }

    private void ensureImeProxy() {
        if (imeProxy != null) {
            return;
        }
        imeProxy = new EditText(this);
        imeProxy.setLayoutParams(new FrameLayout.LayoutParams(1, 1));
        imeProxy.setAlpha(0f);
        imeProxy.setBackgroundColor(Color.TRANSPARENT);
        imeProxy.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_FLAG_MULTI_LINE);
        imeProxy.setImeOptions(EditorInfo.IME_FLAG_NO_FULLSCREEN | EditorInfo.IME_ACTION_NONE);
        imeProxy.setSingleLine(false);
        if (Build.VERSION.SDK_INT >= 21) {
            imeProxy.setShowSoftInputOnFocus(true);
        }
        imeProxy.addTextChangedListener(
                new TextWatcher() {
                    @Override
                    public void beforeTextChanged(CharSequence s, int start, int count, int after) {}

                    @Override
                    public void onTextChanged(CharSequence s, int start, int before, int count) {
                        if (imeSuppress) {
                            return;
                        }
                        try {
                            if (before > 0 && count == 0) {
                                for (int i = 0; i < before; i++) {
                                    nativeImeKey(KeyEvent.KEYCODE_DEL);
                                }
                            }
                            if (count > 0 && s != null) {
                                final int end = Math.min(start + count, s.length());
                                for (int i = start; i < end; ) {
                                    int cp = Character.codePointAt(s, i);
                                    nativeImeCodepoint(cp);
                                    i += Character.charCount(cp);
                                }
                            }
                        } catch (UnsatisfiedLinkError ignored) {
                        }
                        imeSuppress = true;
                        imeProxy.setText("");
                        imeProxy.setSelection(0);
                        imeSuppress = false;
                    }

                    @Override
                    public void afterTextChanged(Editable s) {}
                });
        imeProxy.setOnKeyListener(
                (v, keyCode, event) -> {
                    if (event.getAction() != KeyEvent.ACTION_DOWN) {
                        return false;
                    }
                    if (keyCode == KeyEvent.KEYCODE_DEL) {
                        nativeImeKey(KeyEvent.KEYCODE_DEL);
                        return true;
                    }
                    if (keyCode == KeyEvent.KEYCODE_ENTER) {
                        nativeImeKey(KeyEvent.KEYCODE_ENTER);
                        return true;
                    }
                    return false;
                });
        View decor = getWindow().getDecorView();
        if (decor instanceof ViewGroup) {
            ((ViewGroup) decor).addView(imeProxy);
        } else {
            addContentView(imeProxy, new FrameLayout.LayoutParams(1, 1));
        }
    }

    /** Called from Go via JNI to launch ACTION_OPEN_DOCUMENT. */
    public void gruPickOpenDocument() {
        runOnUiThread(
                () -> {
                    Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
                    intent.addCategory(Intent.CATEGORY_OPENABLE);
                    intent.setType("*/*");
                    intent.putExtra(
                            Intent.EXTRA_MIME_TYPES,
                            new String[] {
                                "text/*",
                                "application/json",
                                "application/javascript",
                                "application/xml",
                                "application/x-yaml",
                                "application/octet-stream"
                            });
                    try {
                        startActivityForResult(intent, REQ_OPEN_DOCUMENT);
                    } catch (Exception e) {
                        deliverOpen("", 1);
                    }
                });
    }

    /**
     * Called from Go via JNI to launch ACTION_CREATE_DOCUMENT. pendingBytes are
     * written when the user confirms; path returned is an app-private copy for
     * subsequent os.WriteFile / re-open.
     */
    public void gruPickSaveDocument(String defaultName, byte[] pendingBytes) {
        pendingSaveName = defaultName != null ? defaultName : "Untitled.txt";
        pendingSaveBytes = pendingBytes != null ? pendingBytes : new byte[0];
        runOnUiThread(
                () -> {
                    Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
                    intent.addCategory(Intent.CATEGORY_OPENABLE);
                    intent.setType("text/plain");
                    intent.putExtra(Intent.EXTRA_TITLE, pendingSaveName);
                    try {
                        startActivityForResult(intent, REQ_CREATE_DOCUMENT);
                    } catch (Exception e) {
                        pendingSaveBytes = null;
                        deliverSave("", 1);
                    }
                });
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode == REQ_OPEN_DOCUMENT) {
            if (resultCode != RESULT_OK || data == null || data.getData() == null) {
                deliverOpen("", 1);
                return;
            }
            handleOpenDocument(data.getData());
            return;
        }
        if (requestCode == REQ_CREATE_DOCUMENT) {
            byte[] bytes = pendingSaveBytes;
            pendingSaveBytes = null;
            if (resultCode != RESULT_OK || data == null || data.getData() == null) {
                deliverSave("", 1);
                return;
            }
            handleCreateDocument(data.getData(), bytes != null ? bytes : new byte[0]);
        }
    }

    private void handleOpenDocument(Uri uri) {
        try {
            takePersistableRead(uri);
            String name = queryDisplayName(uri);
            if (name == null || name.isEmpty()) {
                name = "opened.txt";
            }
            byte[] data = readAll(getContentResolver().openInputStream(uri));
            File out = new File(importsDir(), sanitizeName(name));
            try (FileOutputStream fos = new FileOutputStream(out)) {
                fos.write(data);
            }
            deliverOpen(out.getAbsolutePath(), 0);
        } catch (Exception e) {
            deliverOpen("", 1);
        }
    }

    private void handleCreateDocument(Uri uri, byte[] bytes) {
        try {
            takePersistableWrite(uri);
            try (OutputStream os = getContentResolver().openOutputStream(uri, "wt")) {
                if (os != null) {
                    os.write(bytes);
                    os.flush();
                }
            }
            String name = queryDisplayName(uri);
            if (name == null || name.isEmpty()) {
                name = pendingSaveName != null ? pendingSaveName : "Untitled.txt";
            }
            // Keep a local mirror so Go can os.WriteFile / ReadFile on later Save/Open.
            File out = new File(exportsDir(), sanitizeName(name));
            try (FileOutputStream fos = new FileOutputStream(out)) {
                fos.write(bytes);
            }
            // Remember URI sidecar for future sync-write (best-effort).
            writeUriMap(out.getAbsolutePath(), uri.toString());
            deliverSave(out.getAbsolutePath(), 0);
        } catch (Exception e) {
            deliverSave("", 1);
        }
    }

    private void takePersistableRead(Uri uri) {
        try {
            getContentResolver()
                    .takePersistableUriPermission(
                            uri, Intent.FLAG_GRANT_READ_URI_PERMISSION);
        } catch (SecurityException ignored) {
        }
    }

    private void takePersistableWrite(Uri uri) {
        try {
            getContentResolver()
                    .takePersistableUriPermission(
                            uri,
                            Intent.FLAG_GRANT_READ_URI_PERMISSION
                                    | Intent.FLAG_GRANT_WRITE_URI_PERMISSION);
        } catch (SecurityException ignored) {
        }
    }

    private File importsDir() {
        File d = new File(getFilesDir(), "gru-notepad/imports");
        //noinspection ResultOfMethodCallIgnored
        d.mkdirs();
        return d;
    }

    private File exportsDir() {
        File d = new File(getFilesDir(), "gru-notepad/exports");
        //noinspection ResultOfMethodCallIgnored
        d.mkdirs();
        return d;
    }

    private void writeUriMap(String localPath, String uri) {
        try {
            File map = new File(getFilesDir(), "gru-notepad/uri_map.txt");
            //noinspection ResultOfMethodCallIgnored
            map.getParentFile().mkdirs();
            try (FileOutputStream fos = new FileOutputStream(map, true)) {
                fos.write((localPath + "\t" + uri + "\n").getBytes(StandardCharsets.UTF_8));
            }
        } catch (Exception ignored) {
        }
    }

    private static String sanitizeName(String name) {
        String n = name.replaceAll("[\\\\/:*?\"<>|]", "_");
        if (n.isEmpty()) {
            return "file.txt";
        }
        return n;
    }

    private String queryDisplayName(Uri uri) {
        Cursor c = null;
        try {
            c = getContentResolver().query(uri, null, null, null, null);
            if (c != null && c.moveToFirst()) {
                int idx = c.getColumnIndex(OpenableColumns.DISPLAY_NAME);
                if (idx >= 0) {
                    return c.getString(idx);
                }
            }
        } catch (Exception ignored) {
        } finally {
            if (c != null) {
                c.close();
            }
        }
        return null;
    }

    private static byte[] readAll(InputStream in) throws Exception {
        if (in == null) {
            return new byte[0];
        }
        try (InputStream input = in; ByteArrayOutputStream bos = new ByteArrayOutputStream()) {
            byte[] buf = new byte[8192];
            int n;
            while ((n = input.read(buf)) >= 0) {
                bos.write(buf, 0, n);
            }
            return bos.toByteArray();
        }
    }

    private void deliverOpen(String path, int cancelled) {
        try {
            nativeFileOpenResult(path != null ? path : "", cancelled);
        } catch (UnsatisfiedLinkError ignored) {
        }
    }

    private void deliverSave(String path, int cancelled) {
        try {
            nativeFileSaveResult(path != null ? path : "", cancelled);
        } catch (UnsatisfiedLinkError ignored) {
        }
    }

    private void installImeHeightListener() {
        final View decor = getWindow().getDecorView();
        if (Build.VERSION.SDK_INT >= 30) {
            // Do NOT setDecorFitsSystemWindows(false): that zeros ANativeActivity
            // contentRect and puts chrome under the status bar. Rely on IME
            // animation callbacks + reflection poll under adjustNothing.
            decor.setOnApplyWindowInsetsListener(
                    (v, insets) -> {
                        Insets ime = insets.getInsets(WindowInsets.Type.ime());
                        reportImeHeight(ime.bottom);
                        return v.onApplyWindowInsets(insets);
                    });
            decor.setWindowInsetsAnimationCallback(
                    new WindowInsetsAnimation.Callback(
                            WindowInsetsAnimation.Callback.DISPATCH_MODE_CONTINUE_ON_SUBTREE) {
                        @Override
                        public WindowInsets onProgress(
                                WindowInsets insets, List<WindowInsetsAnimation> runningAnimations) {
                            Insets ime = insets.getInsets(WindowInsets.Type.ime());
                            reportImeHeight(ime.bottom);
                            return insets;
                        }
                    });
        }
        // Always keep a visible-frame / reflection poll — adjustNothing often
        // reports 0 from WindowInsets alone on NativeActivity.
        decor.getViewTreeObserver()
                .addOnGlobalLayoutListener(this::pollImeHeight);
    }

    private void pollImeHeight() {
        int h = 0;
        if (Build.VERSION.SDK_INT >= 30) {
            WindowInsets insets = getWindow().getDecorView().getRootWindowInsets();
            if (insets != null) {
                h = Math.max(h, insets.getInsets(WindowInsets.Type.ime()).bottom);
            }
        }
        if (h <= 0 && imeWanted) {
            h = reflectImeHeight();
        }
        if (h <= 0 && !imeWanted) {
            View decor = getWindow().getDecorView();
            Rect r = new Rect();
            decor.getWindowVisibleDisplayFrame(r);
            h = Math.max(0, decor.getHeight() - r.bottom);
        }
        if (!imeWanted) {
            h = 0;
        }
        reportImeHeight(h);
    }

    private int reflectImeHeight() {
        try {
            InputMethodManager imm =
                    (InputMethodManager) getSystemService(Context.INPUT_METHOD_SERVICE);
            if (imm == null) {
                return 0;
            }
            Method m =
                    InputMethodManager.class.getDeclaredMethod("getInputMethodWindowVisibleHeight");
            m.setAccessible(true);
            Object v = m.invoke(imm);
            if (v instanceof Integer) {
                return Math.max(0, (Integer) v);
            }
        } catch (Exception ignored) {
        }
        return 0;
    }

    private void reportImeHeight(int px) {
        if (px < 0) {
            px = 0;
        }
        if (px == lastImeHeight) {
            return;
        }
        lastImeHeight = px;
        try {
            nativeImeHeight(px);
        } catch (UnsatisfiedLinkError ignored) {
            // Native lib not ready yet during early onCreate.
        }
    }

    private void applyLightSystemBars() {
        Window w = getWindow();
        w.addFlags(WindowManager.LayoutParams.FLAG_DRAWS_SYSTEM_BAR_BACKGROUNDS);
        w.clearFlags(WindowManager.LayoutParams.FLAG_TRANSLUCENT_STATUS);
        w.setStatusBarColor(GRU_BG);
        w.setNavigationBarColor(GRU_BG);
        View decor = w.getDecorView();
        int flags = decor.getSystemUiVisibility();
        flags |= View.SYSTEM_UI_FLAG_LIGHT_STATUS_BAR;
        decor.setSystemUiVisibility(flags);
    }
}
