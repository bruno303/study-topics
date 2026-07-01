# Implementation Plan: Story Backlog Mode

## Overview

Add a story backlog/queue to Planning Poker rooms. The room owner can queue multiple stories, vote on them sequentially, and manually advance between stories. Advance is **always manual** — users must see the result and advance when ready.

## Changes Summary

| Layer | Files Changed | Files Added |
|-------|--------------|-------------|
| Domain | `entity/room.go`, `entity/client.go` | `entity/story.go` |
| Use Cases | `facade.go`, `updatestory.go`, `newvoting.go`, `revealusecase.go`, `voteagain.go` | `togglebacklogmode.go`, `addstory.go`, `removestory.go`, `advancestory.go` |
| DTOs | `dto/commands.go` | — |
| Infra/Bus | `bus/websocketbus.go` | — |
| Infra/Redis | `redis/serialization.go` | — |
| Infra/HTTP | — | — |
| Frontend | `page.tsx`, `page.styles.ts`, `websocket.ts` | — |
| Tests | — | tests for new use cases + entity methods |

## Phase 1: Domain Layer — Entity Changes

### 1.1 New `entity/story.go`

```go
type Story struct {
    Name               string    `json:"name"`
    Result             *float32  `json:"result,omitempty"`
    MostAppearingVotes []int     `json:"mostAppearingVotes"`
    Voted              bool      `json:"voted"`
}
```

### 1.2 Extend `entity/room.go` — Room struct

| Field | Type | Default | Purpose |
|-------|------|---------|---------|
| `BacklogMode` | `bool` | `false` | Toggles backlog mode on/off |
| `Stories` | `[]Story` | `nil` | Ordered list of queued stories |
| `CurrentStoryIndex` | `int` | `0` | Index into Stories (0-based) |

**New methods on `*Room`:**

| Method | Description |
|--------|-------------|
| `ToggleBacklogMode(ctx, clientID)` | Enables/disables backlog mode. When enabling, if a current story exists, add it as first entry. When disabling, set `CurrentStory` to current story name (lose queue). |
| `AddStory(ctx, clientID, name)` | Owner appends a new story to `Stories`. If `BacklogMode` is off and this is the first story, auto-enable backlog mode. |
| `RemoveStory(ctx, clientID, index)` | Owner removes a story by index. Adjust `CurrentStoryIndex` if needed. |
| `AdvanceToNextStory(ctx, clientID)` | Owner manually advances to next un-voted story. Clears votes, hides reveal. If no more stories, stays on last one. |

**Modified existing methods:**

| Method | Change |
|--------|--------|
| `NewVoting()` | In backlog mode, advance `CurrentStoryIndex` to the next story before clearing. In non-backlog mode, keep current behavior (clear `CurrentStory`). |
| `ResetVoting()` | Same as `NewVoting` but does NOT advance. |
| `SetCurrentStory()` | In backlog mode, create/update `Stories[CurrentStoryIndex]`. In non-backlog mode, set `CurrentStory` string (unchanged). |
| `ToggleReveal()` | When revealing in backlog mode, store `Result` and `MostAppearingVotes` into the current `Story`. Mark `Story.Voted = true`. Does NOT auto-advance — advance is always manual. |

### 1.3 Update `entity/room.go` — NewVoting modification

```go
if room.BacklogMode && room.CurrentStoryIndex < len(room.Stories)-1 {
    room.CurrentStoryIndex++
}
```

### 1.4 Update `entity/room.go` — ToggleReveal modification

When revealing with `BacklogMode`:
```go
if reveal && r.BacklogMode && r.CurrentStoryIndex >= 0 && r.CurrentStoryIndex < len(r.Stories) {
    story := &r.Stories[r.CurrentStoryIndex]
    story.Result = r.Result
    story.MostAppearingVotes = r.MostAppearingVotes
    story.Voted = true
}
```

No auto-advance — the room stays on the current story with results visible until the owner manually advances.

---

## Phase 2: Use Case Layer

### 2.1 New use cases

| Use Case | Command | Description |
|----------|---------|-------------|
| `ToggleBacklogMode` | `{RoomID, SenderID}` | Owner toggles backlog mode |
| `AddStory` | `{RoomID, SenderID, StoryName}` | Owner adds a story to backlog |
| `RemoveStory` | `{RoomID, SenderID, StoryIndex}` | Owner removes a story by index |
| `AdvanceStory` | `{RoomID, SenderID}` | Owner manually advances to next story |

Each follows: lock → load room → call domain method → save room → broadcast room state.

### 2.2 Update `facade.go`

```go
ToggleBacklogMode UseCase[ToggleBacklogModeCommand]
AddStory          UseCase[AddStoryCommand]
RemoveStory       UseCase[RemoveStoryCommand]
AdvanceStory      UseCase[AdvanceStoryCommand]
```

### 2.3 Wire in `setup/container.go`

Instantiate and wire all new use cases in `newUsecases()`.

---

## Phase 3: DTO Layer

### 3.1 Update `dto/commands.go` — RoomState

```go
type RoomState struct {
    // ... existing fields ...
    BacklogMode       bool     `json:"backlogMode"`
    Stories           []Story  `json:"stories"`
    CurrentStoryIndex int      `json:"currentStoryIndex"`
}
```

```go
type Story struct {
    Name               string   `json:"name"`
    Result             *float32 `json:"result,omitempty"`
    MostAppearingVotes []int    `json:"mostAppearingVotes"`
    Voted              bool     `json:"voted"`
}
```

---

## Phase 4: Infrastructure — WebSocket Bus

### 4.1 Update `bus/websocketbus.go`

| Message Type | Handler |
|-------------|---------|
| `toggle-backlog-mode` | `ToggleBacklogMode.Execute(...)` |
| `add-story` | `AddStory.Execute(...)` with payload `{story: "..."}` |
| `remove-story` | `RemoveStory.Execute(...)` with payload `{storyIndex: N}` |
| `advance-story` | `AdvanceStory.Execute(...)` |

---

## Phase 5: Infrastructure — Redis Serialization

### 5.1 Update `redis/serialization.go`

```go
type SerializedRoom struct {
    // ... existing fields ...
    BacklogMode       bool               `json:"backlogMode"`
    Stories           []SerializedStory  `json:"stories,omitempty"`
    CurrentStoryIndex int                `json:"currentStoryIndex"`
}
```

```go
type SerializedStory struct {
    Name               string   `json:"name"`
    Result             *float32 `json:"result,omitempty"`
    MostAppearingVotes []int    `json:"mostAppearingVotes"`
    Voted              bool     `json:"voted"`
}
```

---

## Phase 6: Frontend

### 6.1 Update `websocket.ts`

New message types:
```ts
'add-story'
'remove-story'
'toggle-backlog-mode'
'advance-story'
```

Update `RoomState`:
```ts
stories: Array<{ name: string; result?: number; mostAppearingVotes: number[]; voted: boolean }>
backlogMode: boolean
currentStoryIndex: number
```

### 6.2 Update `page.tsx` — Room Page

**New state variables:**
```ts
const [backlogMode, setBacklogMode] = useState(false)
const [stories, setStories] = useState<Story[]>([])
const [currentStoryIndex, setCurrentStoryIndex] = useState(0)
const [newStoryInput, setNewStoryInput] = useState('')
```

**New handlers:**
```ts
handleToggleBacklogMode()
handleAddStory()
handleRemoveStory(index)
handleAdvanceStory()
```

**UI — Backlog Panel (admin only):**

1. **Backlog toggle button** — "Enable Backlog" / "Disable Backlog"
2. **Story list** — Ordered list with: name (editable for current/upcoming), result if voted, status (✅ voted, 🔄 current, ⬜ pending), remove button (pending only)
3. **Add story input** — text input + "Add" button
4. **Next Story button** — manual advance (visible when stories remain)

**Current Story display:**
- In backlog mode, show position: "Story 3 of 5"
- Show previous story results below the card

### 6.3 Update `page.styles.ts`

New styles: `backlogPanel`, `backlogStory`, `backlogStoryCurrent`, `backlogStoryVoted`, `backlogStoryPending`, `backlogInput`, `backlogAddButton`, `backlogToggle`, `nextButton`.

---

## Phase 7: Tests

### 7.1 Backend tests

| Test File | What to Test |
|-----------|-------------|
| `entity/room_test.go` | Backlog mode methods: toggle, add/remove story, advance, reveal stores results |
| `usecase/togglebacklogmode_test.go` | Use case success + error paths |
| `usecase/addstory_test.go` | Adding stories, auto-enabling backlog |
| `usecase/advancestory_test.go` | Manual advance, edge cases (last story) |
| `redis/serialization_test.go` | Serialize/deserialize with backlog fields |
| `dto/commands_test.go` | RoomState mapping with backlog fields |

### 7.2 Frontend tests

- Vitest: Story list rendering, add/remove handlers
- Playwright: Backlog mode E2E flow (enable backlog, add stories, vote, manual advance)

---

## Edge Cases

1. **Backward compatibility**: `BacklogMode=false` → all behavior unchanged. Existing Redis rooms default to zero values.
2. **Empty stories**: Blank current story. Adding first story sets it as current.
3. **Removing current story**: Shift `CurrentStoryIndex` back by 1 (min 0), clear votes.
4. **Last story after voting**: Room stays on last story with results visible.
5. **Exiting backlog mode**: Set `CurrentStory` to current story name, discard queue.
6. **Multiple owners**: Any owner can manage backlog.
7. **Redis TTL**: Stories stored with room (24h TTL). No persistent history.
