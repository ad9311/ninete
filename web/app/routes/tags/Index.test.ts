// @vitest-environment jsdom
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// The page's whole job is turning checkbox picks and a typed name into the
// right request bodies, so those bodies are what these tests assert on.
vi.mock("../../lib/api", async () => {
  const actual =
    await vi.importActual<typeof import("../../lib/api")>("../../lib/api");

  return {
    ...actual,
    get: vi.fn(),
    post: vi.fn(),
    del: vi.fn(),
  };
});

vi.mock("../../router", () => ({ BASE_PATH: "" }));

import { del, get, post } from "../../lib/api";
import Index from "./Index.svelte";

const TAGS = {
  data: [
    { id: 1, name: "taxi", expense_count: 2, recurrent_expense_count: 0 },
    { id: 2, name: "uber", expense_count: 1, recurrent_expense_count: 1 },
    { id: 3, name: "old", expense_count: 0, recurrent_expense_count: 0 },
  ],
};

afterEach(cleanup);
beforeEach(() => {
  vi.mocked(get).mockReset().mockResolvedValue(TAGS);
  vi.mocked(post).mockReset();
  vi.mocked(del).mockReset();
  vi.stubGlobal(
    "confirm",
    vi.fn(() => true),
  );
});

async function checkbox(name: string): Promise<HTMLInputElement> {
  return (await screen.findByRole("checkbox", {
    name: new RegExp(`^${name}\\b`),
  })) as HTMLInputElement;
}

describe("tags page", () => {
  it("shows each tag's usage and flags unused ones", async () => {
    render(Index);

    expect(await screen.findByText("2 expenses")).toBeTruthy();
    expect(screen.getByText("1 expense · 1 recurrent expense")).toBeTruthy();
    expect(screen.getByText("Unused")).toBeTruthy();
    expect(
      screen.getByRole("button", { name: "Delete unused (1)" }),
    ).toBeTruthy();
  });

  it("retags the picked tags onto the typed name", async () => {
    vi.mocked(post).mockResolvedValue({
      tag: { id: 9, name: "transport" },
      retagged: 3,
    });
    render(Index);

    await fireEvent.click(await checkbox("taxi"));
    await fireEvent.click(await checkbox("uber"));
    await fireEvent.input(screen.getByLabelText("Retag as"), {
      target: { value: "transport" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Retag" }));

    expect(post).toHaveBeenCalledWith("/tags/retag", {
      from: ["taxi", "uber"],
      to: "transport",
    });
    expect(
      await screen.findByText(/Moved 3 records onto "transport"/),
    ).toBeTruthy();
  });

  it("keeps Retag disabled until a tag is picked and a name typed", async () => {
    render(Index);
    const button = (await screen.findByRole("button", {
      name: "Retag",
    })) as HTMLButtonElement;

    await waitFor(() => expect(button.disabled).toBe(true));
    await fireEvent.click(await checkbox("taxi"));
    expect(button.disabled).toBe(true);
    await fireEvent.input(screen.getByLabelText("Retag as"), {
      target: { value: "cab" },
    });
    expect(button.disabled).toBe(false);
  });

  it("deletes one tag, telling the owner what it comes off", async () => {
    vi.mocked(del).mockResolvedValue(undefined);
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Delete uber" }),
    );

    expect(vi.mocked(confirm).mock.calls[0][0]).toContain(
      "1 expense and 1 recurrent expense",
    );
    expect(del).toHaveBeenCalledWith("/tags/2");
  });

  it("deletes nothing when the confirm is dismissed", async () => {
    vi.mocked(confirm).mockReturnValue(false);
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Delete unused (1)" }),
    );

    expect(del).not.toHaveBeenCalled();
  });

  it("deletes every unused tag in one request", async () => {
    vi.mocked(del).mockResolvedValue({ deleted: 1 });
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Delete unused (1)" }),
    );

    expect(del).toHaveBeenCalledWith("/tags/unused");
    expect(await screen.findByText("Deleted 1 unused tag.")).toBeTruthy();
  });
});
