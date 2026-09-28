champ_push() {
  local file="$1"
  local name=$(basename "$file")
  local size=$(stat -c%%s "$file")
  local data=$(base64 -w0 "$file")

  [ -n "$TMUX" ] && tmux set -g allow-passthrough

  printf "\033]1337;File=name=%%s;size=%%s;host=%%s;directory=%%s;inline=0:%%s\a" "$(echo -n "$name" | base64 -w0)" "$size" $(hostname -f) "true" "$data"

  [ -n "$TMUX" ] && tmux refresh-client
};

export -f champ_push

exec /bin/bash --login

# TODO: alle über gebliebenen Ptmux entfernen am ende
# sudo alias
# XDG_CONFIG_HOME
