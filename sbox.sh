#!/usr/bin/env bash
set -euo pipefail

if [ $# -eq 0 ]; then
  echo "Usage: $0 <claude|agy|command> [arguments...]"
  exit 1
fi

CMD="$1"
shift

EXTRA_ARGS=()
case "$CMD" in
  claude|agy)
    # Ajoute --dangerously-skip-permissions si non spécifié explicitement
    if [[ ! " $* " =~ " --dangerously-skip-permissions " ]]; then
      EXTRA_ARGS+=(--dangerously-skip-permissions)
    fi
    ;;
esac

# Liste des répertoires et fichiers de configuration à exposer
CONFIG_MOUNTS=()
for path in \
  "$HOME/.claude" \
  "$HOME/.claude.json" \
  "$HOME/.config/claude" \
  "$HOME/.gemini" \
  "$HOME/.config/agy" \
  "$HOME/.config/herdr" \
  "$HOME/.gitconfig"; do
  CONFIG_MOUNTS+=(--bind-try "$path" "$path")
done

# Sockets Wayland & Quickshell
XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"
WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-0}"
WAYLAND_SOCKET="${XDG_RUNTIME_DIR}/${WAYLAND_DISPLAY}"

exec bwrap \
  --ro-bind /usr /usr \
  --ro-bind /lib /lib \
  --ro-bind /lib64 /lib64 \
  --ro-bind /bin /bin \
  --ro-bind /etc /etc \
  --ro-bind-try /opt /opt \
  --ro-bind-try /run/systemd/resolve /run/systemd/resolve \
  --proc /proc \
  --dev /dev \
  --tmpfs /tmp \
  --bind-try "$WAYLAND_SOCKET" "$WAYLAND_SOCKET" \
  --bind-try "$XDG_RUNTIME_DIR/quickshell" "$XDG_RUNTIME_DIR/quickshell" \
  --bind-try "$XDG_RUNTIME_DIR/hypr" "$XDG_RUNTIME_DIR/hypr" \
  --dir "$HOME" \
  "${CONFIG_MOUNTS[@]}" \
  --bind "$PWD" "$PWD" \
  --chdir "$PWD" \
  --die-with-parent \
  --new-session \
  -- "$CMD" "${EXTRA_ARGS[@]}" "$@"
