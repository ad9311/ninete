import { get } from "./api";

// Categories are a shared lookup table (CLAUDE.md), not a resource of their
// own — every form that references one needs only this much of it.
export interface Category {
  id: number;
  name: string;
}

interface CategoryListResponse {
  data: Category[];
}

export async function fetchCategories(): Promise<Category[]> {
  const response = await get<CategoryListResponse>("/categories");

  return response.data;
}
