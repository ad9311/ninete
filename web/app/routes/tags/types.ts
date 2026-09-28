// Mirrors handle_api_tags.go's JSON shape exactly, snake_case included
// (§3.5 of docs/spa-migration.md).

export interface TagUsage {
  id: number;
  name: string;
  expense_count: number;
  recurrent_expense_count: number;
}

export interface TagListResponse {
  data: TagUsage[];
}

export interface RetagResponse {
  tag: { id: number; name: string };
  /** Distinct expenses and recurrent expenses moved onto `tag`. */
  retagged: number;
}

export interface DeleteUnusedTagsResponse {
  deleted: number;
}
