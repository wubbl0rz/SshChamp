_champ_init() {
  # max 10MB for clipboard
  export CHAMP_MAX_CLIPBOARD_SIZE=10485760
  export CHAMP_HAS_CLIP='{{.CHAMP_HAS_CLIP}}'

  for f in $(compgen -A function | grep -E '^_?champ_'); do
    export -f "${f?}"
    if [ "$1" = "--show" ]; then
      declare -f "${f?}"
    fi
  done
}

_champ_err() { printf 'error: %s\n' "$*" >&2; return 255; }

champ_push() (
  usage() {
    cat <<'EOF'
Usage: champ_push <file>  -  send file to local machine

Example:
  champ_push ./report.pdf
  echo 123 | champ_push
EOF
  }

  check_input() (
    local data=$1 size=$2

    if [ "$size" -gt "$CHAMP_MAX_CLIPBOARD_SIZE" ]; then
      _champ_err "data is too big for the clipboard (>$(awk -v bytes="$CHAMP_MAX_CLIPBOARD_SIZE" 'BEGIN { printf "%.0f MiB\n", bytes / 1048576 }'))"
      return 255
    fi

    if ! printf '%s' "$data" | base64 -d | grep -Iq '.'; then
      _champ_err "binary data in input. this might break your clipboard"
      return 255
    fi
  )

  if { { [ -z "$1" ] || [ "$1" = "-" ]; } && [ -t 0 ]; } || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  local target="file"

  if [ "$2" = "--clip" ]; then
    if ! "$CHAMP_HAS_CLIP"; then
      _champ_err "clipboard on local machine not available"
      return 255
    fi
    target="clip"
  fi

  if [ ! -t 0 ]; then
    local name data size

    name="stdin"
    data=$(base64 -w0)
    size=$(printf '%s' "$data" | base64 -d | wc -c)

    if [ "$target" = "clip" ]; then
      check_input "$data" "$size" || { return 255; }
    fi
  else
    local source="$1"

    if [ ! -e "$source" ]; then
      _champ_err "file or directory ($1) does not exist"
      return 255
    fi

    local name size data is_directory
    is_directory=false

    if [ -d "$source" ]; then
      is_directory=true
      name="$(basename -- "$source").tar"
      data=$(tar -C "$source" -cf - . | base64 -w0) || {
        _champ_err "could not archive directory ($source)"
        return 255
      }
    elif [ -f "$source" ]; then
      name=$(basename -- "$source")
      data=$(base64 -w0 < "$source") || {
        _champ_err "could not read file ($source)"
        return 255
      }
    else
      _champ_err "not a regular file or directory ($source)"
      return 255
    fi

    size=$(printf '%s' "$data" | base64 -d | wc -c)

    if "$is_directory" && [ "$target" = "clip" ]; then
      _champ_err "directory cannot be copied to clipboard ($source)"
      return 255
    fi

    if [ "$target" = "clip" ]; then
      check_input "$data" "$size" || return 255
    fi
  fi

  [ -n "$TMUX" ] && tmux set -g allow-passthrough on

  printf "\033]1337;File=name=%s;size=%s;host=%s;target=%s;directory=%s;inline=0:%s\a" "$(echo -n "$name" | base64 -w0)" "$size" "$(hostname -f)" "$target" "$is_directory" "$data"

  [ -n "$TMUX" ] && tmux refresh-client
)

champ_clip() (
  usage() {
    cat <<'EOF'
Usage: champ_clip <file>  -  send file to local clipboard

Example:
  champ_clip ./report.pdf
  echo 123 | champ_clip
EOF
  }

  if  { [ -z "$1" ] && [ -t 0 ]; } || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  if [ ! -t 0 ]; then
    champ_push "-" "--clip"
  else
    champ_push "$1" "--clip"
  fi
);

champ_sudo() (
  usage() {
    cat <<'EOF'
Usage: champ_sudo <user>  -  sudo wrapper to keep all champ_* functions

Example:
  # become root
  champ_sudo
  # switch user
  champ_sudo <user>
EOF
  }

  if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  sudo -u "${1:-root}" bash -c "$(_champ_init --show); _champ_init; exec bash -l"
)

champ_ssh() (
  usage() {
    cat <<'EOF'
Usage: champ_ssh <user@host>  -  ssh wrapper to keep all champ_* functions
EOF
  }

  if [ -z "$1" ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  ssh -t "$@" "/bin/bash -c $(printf "%q" "$(_champ_init --show); _champ_init; exec bash -l")"
)

champ_save() (
  local file_name="./champ_init.sh"

  usage() {
    cat <<EOF
Usage: champ_save  -  save all champ_* functions to a script ($file_name)
EOF
  }

  if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    usage
    return 0
  fi

  echo "#!/bin/bash" > "$file_name"
  _champ_init --show >> "$file_name"
  echo "_champ_init; exec /bin/bash --login" >> "$file_name"

  chmod +x "$file_name"
)

_champ_init

exec /bin/bash --login

# todo serverseite fragen ob unzip oder nicht
# command to create scripts that can be passed to containers for example
# TODO option for copy base64 to clipboard if binary
# todo: alles in main func und nur die exporten ? <--- ne dann exportet der ja net die _* functions
# todo edit command
#todo support directories
# todo support multiple files and *
# TODO tmux passtrough
# TODO clipboard handling
# TODO: alle über gebliebenen Ptmux entfernen am ende
# XDG_CONFIG_HOME
