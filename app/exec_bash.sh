champ_push() {
  usage() {
    cat <<'EOF'
Usage: champ_push <file>  -  send file to local machine

Example:
  champ_push ./report.pdf
  echo 123 | champ_push
EOF
  }

  check_input() {
    if ! printf '%%s' "$data" | base64 -d | grep -Iq '.'; then
      echo -e "\n\nerror: binary data in input. this might break your clipboard.\n\n"
      return 255
    fi

    # max 10MB for clipboard
    if [ "$size" -gt 10485760 ]; then
      echo -e "\n\nerror: data is really big for clipboard (>10MB).\n\n"
      return 255
    fi
  }

  if (([ -z "$1" ] || [ "$1" = "-" ] ) && [ -t 0 ]) || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  local target="file"

  if [ "$2" = "--clip" ]; then
    #TODO
    #if ! $CHAMP_HAS_CLIP; then
    if false; then
      echo "error: clipboard on local machine not available"
      return 255
    fi
    target="clip"
  fi

  local escape_output="true"

  if [ "$3" = "-R" ]; then
    escape_output="false"
  fi

  # TODO option for copy base64 to clipboard if binary
  # TODO accept raw input -R
  # todo -R für not escape binary input
  # oder besser -R für stream input
  if [ ! -t 0 ]; then
    local name="stdin"
    local data=$(tee >(cat >&2) | base64 -w0)
    local size=$(printf '%%s' "$data" | base64 -d | wc -c)

    if [ "$target" = "clip" ]; then
      check_input "$data" "$size" || { return 255; }
    fi
  else
    if [ ! -f "$1" ]; then
      echo "error: $1 does not exist"
      return 255
    fi

    local file="$1"
    local name=$(basename "$file")
    local size=$(stat -c%%s "$file")
    local data=$(base64 -w0 "$file")

    if [ "$target" = "clip" ]; then
      check_input "$data" "$size" || return 255
    fi
  fi

  [ -n "$TMUX" ] && tmux set -g allow-passthrough

  printf "\033]1337;File=name=%%s;size=%%s;host=%%s;target=%%s;directory=%%s;inline=0:%%s\a" "$(echo -n "$name" | base64 -w0)" "$size" $(hostname -f) "$target" "true" "$data"

  [ -n "$TMUX" ] && tmux refresh-client
};

champ_clip() {
  usage() {
    cat <<'EOF'
Usage: champ_clip <file>  -  send file to local clipboard

Example:
  champ_clip ./report.pdf
  echo 123 | champ_clip
EOF
  }

  if ([ -z "$1" ] && [ -t 0 ]) || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  if [ ! -t 0 ]; then
    champ_push "-" "--clip"
  else
    champ_push "$1" "--clip"
  fi
};

export -f champ_push
export -f champ_clip

exec /bin/bash --login

#todo support directories
# todo support multiple files and *
# todo handle stdin
# TODO tmux passtrough
# TODO clipboard handling
# TODO: alle über gebliebenen Ptmux entfernen am ende
# sudo alias
# XDG_CONFIG_HOME
