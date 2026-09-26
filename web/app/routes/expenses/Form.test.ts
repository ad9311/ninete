// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";

// Covers the note field's client-side half: the counter counts what the
// server will validate (characters after trimming), an over-long note is
// stopped before the request, and the body carries the trimmed note.
vi.mock("../../lib/categories", () => ({
  fetchCategories: vi.fn(() => Promise.resolve([])),
}));

import Form from "./Form.svelte";
import { noteLength, type ExpenseDetail } from "./types";

afterEach(cleanup);

const initial: ExpenseDetail = {
  id: 1,
  category_id: 1,
  category_name: "Food",
  description: "Groceries",
  amount: 2599,
  date: 1755993600,
  created_at: 1755993600,
  tags: [],
  note: "",
};

function noteField(): HTMLTextAreaElement {
  return screen.getByLabelText("Note") as HTMLTextAreaElement;
}

describe("noteLength", () => {
  it("counts characters, not UTF-16 units, after trimming", () => {
    expect(noteLength("  abc  ")).toBe(3);
    expect(noteLength("🙂🙂")).toBe(2);
    expect(noteLength("   ")).toBe(0);
  });
});

describe("expense form note", () => {
  it("seeds the note and submits it trimmed", async () => {
    const onSubmit = vi.fn();
    render(Form, {
      initial: { ...initial, note: "old" },
      submitLabel: "Save",
      onSubmit,
    });

    expect(noteField().value).toBe("old");
    await fireEvent.input(noteField(), { target: { value: "  new note  " } });
    expect(screen.getByText("8/255")).toBeTruthy();

    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(onSubmit).toHaveBeenCalledOnce();
    expect(onSubmit.mock.calls[0][0].note).toBe("new note");
  });

  it("blocks an over-long note before submitting", async () => {
    const onSubmit = vi.fn();
    render(Form, { initial, submitLabel: "Save", onSubmit });

    await fireEvent.input(noteField(), {
      target: { value: "n".repeat(256) },
    });
    expect(screen.getByText("256/255").classList).toContain("text-danger");

    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(onSubmit).not.toHaveBeenCalled();
    expect(
      screen.getByText("Note must be at most 255 characters."),
    ).toBeTruthy();
  });
});
