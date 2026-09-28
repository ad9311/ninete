// @vitest-environment jsdom
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// The page's whole job is turning a row's rename, the ticked rows' merge and
// the deletes into the right requests, so those requests are what these tests
// assert on.
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

  it("renames one tag from its row", async () => {
    vi.mocked(post).mockResolvedValue({
      tag: { id: 9, name: "cab" },
      retagged: 2,
    });
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Rename taxi" }),
    );
    await fireEvent.input(screen.getByLabelText("New name for taxi"), {
      target: { value: "cab" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(post).toHaveBeenCalledWith("/tags/retag", {
      from: ["taxi"],
      to: "cab",
    });
    expect(await screen.findByText(/Moved 2 records onto "cab"/)).toBeTruthy();
    await waitFor(() => {
      expect(screen.queryByLabelText("New name for taxi")).toBeNull();
    });
  });

  it("warns before a rename that merges into an existing tag", async () => {
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Rename taxi" }),
    );
    await fireEvent.input(screen.getByLabelText("New name for taxi"), {
      target: { value: " Uber " },
    });

    expect(screen.getByText(/"uber" already exists/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "Merge" })).toBeTruthy();
  });

  it("keeps Save disabled for an empty or unchanged name", async () => {
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Rename taxi" }),
    );
    const save = screen.getByRole("button", {
      name: "Save",
    }) as HTMLButtonElement;
    expect(save.disabled).toBe(true);

    await fireEvent.input(screen.getByLabelText("New name for taxi"), {
      target: { value: "  " },
    });
    expect(save.disabled).toBe(true);
  });

  it("shows the merge bar only once two tags are ticked", async () => {
    render(Index);

    await fireEvent.click(await checkbox("taxi"));
    expect(screen.queryByText(/Merge 1 tags into/)).toBeNull();

    await fireEvent.click(await checkbox("uber"));
    expect(screen.getByText("Merge 2 tags into")).toBeTruthy();
  });

  it("merges the ticked tags onto the typed name", async () => {
    vi.mocked(post).mockResolvedValue({
      tag: { id: 9, name: "transport" },
      retagged: 3,
    });
    render(Index);

    await fireEvent.click(await checkbox("taxi"));
    await fireEvent.click(await checkbox("uber"));
    await fireEvent.input(screen.getByLabelText(/Merge 2 tags into/), {
      target: { value: "transport" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Merge" }));

    expect(post).toHaveBeenCalledWith("/tags/retag", {
      from: ["taxi", "uber"],
      to: "transport",
    });
    expect(
      await screen.findByText(/Moved 3 records onto "transport"/),
    ).toBeTruthy();
  });

  it("keeps a ticked tag that is also the merge target", async () => {
    vi.mocked(post).mockResolvedValue({
      tag: { id: 1, name: "taxi" },
      retagged: 2,
    });
    render(Index);

    await fireEvent.click(await checkbox("taxi"));
    await fireEvent.click(await checkbox("uber"));
    await fireEvent.input(screen.getByLabelText(/Merge 2 tags into/), {
      target: { value: "Taxi" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Merge" }));

    expect(post).toHaveBeenCalledWith("/tags/retag", {
      from: ["uber"],
      to: "Taxi",
    });
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

  it("clears a failed reload's error once a later reload succeeds", async () => {
    vi.mocked(del).mockResolvedValue(undefined);
    vi.mocked(get)
      .mockResolvedValueOnce(TAGS)
      .mockRejectedValueOnce(new Error("offline"))
      .mockResolvedValue(TAGS);
    render(Index);

    const button = await screen.findByRole("button", { name: "Delete uber" });
    await fireEvent.click(button);
    expect(await screen.findByText("Something went wrong.")).toBeTruthy();

    await fireEvent.click(button);
    await waitFor(() => {
      expect(screen.queryByText("Something went wrong.")).toBeNull();
    });
  });

  it("clears a failed delete's error when a rename then succeeds", async () => {
    vi.mocked(del).mockRejectedValue(new Error("offline"));
    vi.mocked(post).mockResolvedValue({
      tag: { id: 9, name: "cab" },
      retagged: 2,
    });
    render(Index);

    await fireEvent.click(
      await screen.findByRole("button", { name: "Delete uber" }),
    );
    expect(await screen.findByText("Something went wrong.")).toBeTruthy();

    await fireEvent.click(screen.getByRole("button", { name: "Rename taxi" }));
    await fireEvent.input(screen.getByLabelText("New name for taxi"), {
      target: { value: "cab" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText(/Moved 2 records onto "cab"/)).toBeTruthy();
    expect(screen.queryByText("Something went wrong.")).toBeNull();
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
