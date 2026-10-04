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

  # TODO option for copy base64 to clipboard if binary
  if [ ! -t 0 ]; then
    local name="stdin"
    local data=$(base64 -w0)
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

champ_sudo() {
  usage() {
    cat <<'EOF'
Usage: champ_sudo <user>  -  sudo wrapper to keep all champ_* functions

Example:
  # become root
  champ_sudo
  # switch user
  champ_sudo user
EOF
  }

  if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  if [ -z "$1" ]; then
    sudo -u root bash -c "$(declare -f champ_push champ_clip champ_sudo); export -f champ_push champ_clip champ_sudo; exec bash -l"
    return 0
  fi
  sudo -u "$1" bash -c "$(declare -f champ_push champ_clip champ_sudo); export -f champ_push champ_clip champ_sudo; exec bash -l"
}

champ_ssh() {
  usage() {
    cat <<'EOF'
Usage: champ_ssh <user@host>  -  ssh wrapper to keep all champ_* functions
EOF
  }

  if [ -z "$1" ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  local remote_cmd
  remote_cmd="$(declare -f champ_push champ_clip champ_sudo champ_ssh)"
  remote_cmd+="; export -f champ_push champ_clip champ_sudo champ_ssh"
  remote_cmd+="; exec bash -l"

  ssh -t "$@" "/bin/bash -c $(printf "%%q" "$remote_cmd")"
}

export -f champ_push
export -f champ_clip
export -f champ_sudo
export -f champ_ssh

exec /bin/bash --login

#todo support directories
# todo support multiple files and *
# todo handle stdin
# TODO tmux passtrough
# TODO clipboard handling
# TODO: alle über gebliebenen Ptmux entfernen am ende
# sudo alias
# XDG_CONFIG_HOME
