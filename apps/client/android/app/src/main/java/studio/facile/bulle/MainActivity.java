package studio.facile.bulle;

import android.content.res.Configuration;
import android.view.Window;
import androidx.core.content.ContextCompat;
import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsControllerCompat;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {

    /**
     * Repaints the system bars when the device switches between light and dark mode,
     * so the bars keep the page color and readable icons without recreating the WebView.
     */
    @Override
    public void onConfigurationChanged(Configuration newConfig) {
        super.onConfigurationChanged(newConfig);
        boolean night = (newConfig.uiMode & Configuration.UI_MODE_NIGHT_MASK) == Configuration.UI_MODE_NIGHT_YES;
        Window window = getWindow();
        WindowInsetsControllerCompat controller = WindowCompat.getInsetsController(window, window.getDecorView());
        controller.setAppearanceLightStatusBars(!night);
        controller.setAppearanceLightNavigationBars(!night);
        int page = ContextCompat.getColor(createConfigurationContext(newConfig), R.color.bulle_page);
        window.getDecorView().setBackgroundColor(page);
        window.setStatusBarColor(page);
        window.setNavigationBarColor(page);
    }
}
