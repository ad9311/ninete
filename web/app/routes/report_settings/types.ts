// Mirrors handle_api_report_settings.go's JSON shape exactly, snake_case
// included (§3.5 of docs/spa-migration.md).

export interface ReportTag {
  id: number;
  name: string;
}

export interface ReportSettingsResponse {
  selected_tag_ids: number[];
  tags: ReportTag[];
  /** How many grouping tags the server accepts (logic.ReportTagLimit). */
  tag_limit: number;
}
