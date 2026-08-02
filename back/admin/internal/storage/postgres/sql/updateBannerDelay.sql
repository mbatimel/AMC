UPDATE portal_banner_settings
SET delay_sec = $1, updated_at = now()
WHERE singleton = TRUE;
