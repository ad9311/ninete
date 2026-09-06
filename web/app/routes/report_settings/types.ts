// Mirrors handle_api_report_settings.go's JSON shape exactly, snake_case
// included (§3.5 of docs/spa-migration.md).

export interface ReportTag {
  id: number;
  name: string;
}

export interface ReportSettingsResponse {
  timezone: string;
  /** False until the settings have been saved once, so the form seeds its
   * timezone from the browser instead of showing the server's UTC fallback. */
  configured: boolean;
  selected_tag_ids: number[];
  tags: ReportTag[];
}
