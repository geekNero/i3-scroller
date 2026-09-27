# i3-scroller: Infinite Workspace Ribbon

This Go daemon implements an infinite, horizontally scrolling viewport for i3wm. It bypasses the standard spatial limitations of static tiling by utilizing hidden auxiliary workspaces as Last-In-First-Out (LIFO) stacks. This allows users to push windows off-screen and seamlessly pull them back without breaking native container layouts.

## CLI Commands & Features

The binary accepts a single argument to manipulate the active workspace's root-level containers.

*   `scroll-left` / `scroll-right`: **(Push 1, Pull 1)** Pushes the container on the trailing edge into the corresponding hidden history stack, and simultaneously pulls the top container from the opposite history stack into the leading edge. Focus is shifted to the newly pulled window.
*   `move-left` / `move-right`: **(Push 1, Pull 0)** Ejects the container on the specified edge of the current workspace into the hidden history stack. It does not pull a replacement, effectively clearing space on the screen.
*   `collect-left` / `collect-right`: **(Push 0, Pull 1)** Retrieves the most recently pushed container from the specified hidden stack and places it on the corresponding edge of the active workspace.

## Core Architectural Mechanics

### 1. Polybar Evasion (`getSpelledName`)
To prevent the hidden history stacks from cluttering Polybar or other pagers, the daemon dynamically translates workspace integers into alphabetical strings. If the active workspace is `1`, the hidden storage workspaces are instantiated as `LOne` and `ROne`. Polybar ignores these non-integer workspaces unless explicitly configured otherwise, keeping the UI clean.

### 2. Stateless Tree Traversal (`getEdgeNode`)
The daemon holds zero state in memory. It relies entirely on the live i3 IPC tree (`go.i3wm.org/i3/v4`). By traversing the `Nodes` array of a workspace, it dynamically identifies the `0` index (absolute left) and `len-1` index (absolute right). This makes the tool immune to spontaneous application crashes or manual window closures.

### 3. Atomic Placement and Split Escaping (`moveContainerToWorkspace`)
Moving containers precisely in i3 without them collapsing into existing splits requires a highly orchestrated chain of commands, executed atomically to prevent X11 visual tearing.
*   **Targeting:** The script uses i3 marks to identify the current edge container, moves the incoming container to that mark, and instantly unmarks it.
*   **Directional Swapping:** Moving to a mark inherently places the incoming window after (to the right of) the target. If the intended edge is LEFT, the script appends a move left command to swap their positions.
*   **Split Escape Hatch:** If the target edge container is a layout split (identified by `markContainer.Name == ""` since only leaf windows have title names), i3 natively absorbs the incoming window inside that split. The script detects this missing name and appends a final move left or move right command, forcing the incoming window to pop out of the nested layout and become a true root-level sibling.

## Usage Setup

Compile the binary and map the commands directly to your i3 configuration file (`~/.config/i3/config`):

```bash
# Scroll the viewport
bindsym $mod+Right exec /path/to/i3-scroller scroll-right
bindsym $mod+Left  exec /path/to/i3-scroller scroll-left

# Eject a window off-screen
bindsym $mod+Shift+Right exec /path/to/i3-scroller move-right
bindsym $mod+Shift+Left  exec /path/to/i3-scroller move-left

# Retrieve a window from off-screen
bindsym $mod+Control+Right exec /path/to/i3-scroller collect-right
bindsym $mod+Control+Left  exec /path/to/i3-scroller collect-left
```

*Note*: The script is handwritten by me, except this README.md 
