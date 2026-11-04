#!/bin/sh
# First-party bootstrap. The versioned release installer owns binary/tool setup.
# Execution stays at the end so a truncated function cannot partially execute.

l_install_fail() { printf 'L setup failed: %s\n' "$*" >&2; exit 1; }

l_install_path() (
    newline=$(printf '\nx'); newline=${newline%x}
    carriage=$(printf '\r')
    case $1 in
        ''|*:*|*"$newline"*|*"$carriage"*) l_install_fail 'Installation paths must be nonempty and contain no colons or line breaks.' ;;
    esac
)

# Resolve links without replacing them. Missing final components are allowed.
l_install_resolve() (
    l_install_path "$1" || exit 1
    case $1 in /*) rest=${1#/} ;; *) rest=${PWD#/}/$1 ;; esac
    resolved=/
    links=0
    while [ -n "$rest" ]; do
        part=${rest%%/*}
        case $rest in */*) rest=${rest#*/} ;; *) rest= ;; esac
        case $part in ''|.) continue ;; ..) resolved=$(dirname "$resolved"); continue ;; esac
        next=${resolved%/}/$part
        if [ -L "$next" ]; then
            links=$((links + 1))
            [ "$links" -le 40 ] || l_install_fail "Too many symbolic links: $1"
            [ -e "$next" ] || l_install_fail "Broken symbolic link: $next"
            link=$(readlink "$next") || exit 1
            l_install_path "$link" || exit 1
            case $link in /*) next=$link ;; *) next=${resolved%/}/$link ;; esac
            rest=${next#/}${rest:+/$rest}
            resolved=/
        else
            if [ -n "$rest" ] && [ -e "$next" ] && [ ! -d "$next" ]; then
                l_install_fail "Not a directory: $next"
            fi
            resolved=$next
        fi
    done
    printf '%s\n' "$resolved"
)

l_install_writable() (
    if [ -e "$1" ]; then
        [ -f "$1" ] && [ -w "$1" ] || l_install_fail "Not a writable regular file: $1"
    fi
    parent=$(dirname "$1")
    while [ ! -e "$parent" ]; do parent=$(dirname "$parent"); done
    [ -d "$parent" ] && [ -w "$parent" ] || l_install_fail "Directory is not writable: $parent"
)

l_install_quote() {
    printf "'"
    printf '%s' "$1" | sed "s/'/'\\\\''/g"
    printf "'"
}

l_install_mode() { ls -ld "$1" | awk '{print $1}'; }

l_install_profile() (
    original=$1
    # Byte offsets preserve everything outside the owned block, including a
    # final line without a newline. LC_ALL=C makes awk lengths byte lengths.
    offsets=$(awk '
      BEGIN { start=-1; end=-1; active=0; bad=0; pos=0 }
      $0 == "# >>> L environment >>>" {
        if (active || start >= 0) bad=1;
        start=pos; active=1;
      }
      $0 == "# <<< L environment <<<" {
        if (!active) bad=1;
        end=pos+length($0)+1; active=0;
      }
      index($0, "# >>> L environment") == 1 && $0 != "# >>> L environment >>>" { bad=1 }
      index($0, "# <<< L environment") == 1 && $0 != "# <<< L environment <<<" { bad=1 }
      { pos+=length($0)+1 }
      END { if (bad || active) exit 1; print start, end }
    ' "$original") || l_install_fail "Malformed or duplicate L environment block in $2"
    set -- $offsets
    begin=$1
    end=$2
    if [ "$begin" -gt 0 ]; then
        head -c "$begin" "$original"
    elif [ "$begin" -lt 0 ]; then
        cat "$original"
        if [ -s "$original" ] && [ -n "$(tail -c 1 "$original")" ]; then printf '\n'; fi
    fi
    cat <<'L_PROFILE'
# >>> L environment >>>
[ ! -r "$HOME/.local/share/l/env" ] || . "$HOME/.local/share/l/env"
# <<< L environment <<<
L_PROFILE
    if [ "$end" -ge 0 ]; then tail -c "+$((end + 1))" "$original"; fi
)

l_install_environment() {
    printf '%s\n' '# L shell environment (managed by the L installer)'
    # Persist custom locations only as fallbacks. Explicit overrides still win.
    if [ -n "${CARGO_HOME:-}" ]; then
        printf 'if [ -z "${CARGO_HOME:-}" ]; then CARGO_HOME='
        l_install_quote "$l_install_cargo"
        printf '; export CARGO_HOME; fi\n'
    fi
    if [ -n "${RUSTUP_HOME:-}" ]; then
        printf 'if [ -z "${RUSTUP_HOME:-}" ]; then RUSTUP_HOME='
        l_install_quote "$l_install_rustup"
        printf '; export RUSTUP_HOME; fi\n'
    fi
    printf '_l_path_prefix='
    l_install_quote "$l_install_prefix"
    printf '\n'
    cat <<'L_ENV'
_l_path_cargo="${CARGO_HOME:-$HOME/.cargo}/bin"
_l_path_rest=${PATH-}
_l_path_kept=
_l_path_separator=
while :; do
    case $_l_path_rest in
        *:*) _l_path_entry=${_l_path_rest%%:*}; _l_path_rest=${_l_path_rest#*:}; _l_path_more=yes ;;
        *) _l_path_entry=$_l_path_rest; _l_path_more=no ;;
    esac
    if [ "$_l_path_entry" != "$_l_path_prefix" ] && [ "$_l_path_entry" != "$_l_path_cargo" ]; then
        _l_path_kept=$_l_path_kept$_l_path_separator$_l_path_entry
        _l_path_separator=:
    fi
    [ "$_l_path_more" = yes ] || break
done
PATH=$_l_path_prefix
[ "$_l_path_prefix" = "$_l_path_cargo" ] || PATH=$PATH:$_l_path_cargo
[ -z "$_l_path_separator" ] || PATH=$PATH:$_l_path_kept
export PATH
unset _l_path_prefix _l_path_cargo _l_path_rest _l_path_kept _l_path_separator _l_path_entry _l_path_more
L_ENV
}

# Snapshot each logical path and stage edits to its resolved regular-file target.
l_install_stage() {
    l_install_logical=$1
    l_install_kind=$2
    l_install_target=$(l_install_resolve "$l_install_logical") || exit 1
    l_install_writable "$l_install_target"
    l_install_i=1
    while [ "$l_install_i" -le "$l_install_count" ]; do
        l_install_entry=$l_install_work/state/$l_install_i
        if [ "$(cat "$l_install_entry/target")" = "$l_install_target" ]; then
            [ "$(cat "$l_install_entry/kind")" = "$l_install_kind" ] || l_install_fail 'A startup file points to the L environment file.'
            printf '%s\n' "$l_install_logical" >> "$l_install_entry/logical"
            return
        fi
        l_install_i=$((l_install_i + 1))
    done
    l_install_count=$((l_install_count + 1))
    l_install_entry=$l_install_work/state/$l_install_count
    mkdir "$l_install_entry"
    printf '%s\n' "$l_install_logical" > "$l_install_entry/logical"
    printf '%s\n' "$l_install_target" > "$l_install_entry/target"
    printf '%s\n' "$l_install_kind" > "$l_install_entry/kind"
    if [ -e "$l_install_target" ]; then
        l_install_mode "$l_install_target" > "$l_install_entry/mode"
        cp -p "$l_install_target" "$l_install_entry/before"
        : > "$l_install_entry/existed"
    else
        : > "$l_install_entry/before"
    fi
    if [ "$l_install_kind" = env ]; then
        if [ -f "$l_install_entry/existed" ]; then
            [ "$(head -n 1 "$l_install_entry/before")" = '# L shell environment (managed by the L installer)' ] ||
                l_install_fail "Refusing to overwrite an unrecognized environment file: $l_install_target"
        fi
        l_install_environment > "$l_install_entry/content"
    else
        l_install_profile "$l_install_entry/before" "$l_install_logical" > "$l_install_entry/content"
    fi
    cp -p "$l_install_entry/before" "$l_install_entry/after"
    cat "$l_install_entry/content" > "$l_install_entry/after"
}

l_install_unchanged() (
    entry=$1
    target=$(cat "$entry/target")
    while IFS= read -r logical; do
        [ "$(l_install_resolve "$logical")" = "$target" ] ||
            l_install_fail "Startup path changed during setup: $logical"
    done < "$entry/logical"
    if [ -f "$entry/existed" ]; then
        [ -f "$target" ] && cmp -s "$target" "$entry/before" ||
            l_install_fail "Configuration changed during setup: $target"
        [ "$(l_install_mode "$target")" = "$(cat "$entry/mode")" ] ||
            l_install_fail "Configuration permissions changed during setup: $target"
    else
        [ ! -e "$target" ] && [ ! -L "$target" ] ||
            l_install_fail "Configuration appeared during setup: $target"
    fi
)

l_install_replace() {
    l_install_replace_temp=$(mktemp "$(dirname "$2")/.l-install.XXXXXXXX") || return 1
    if cp -p "$1" "$l_install_replace_temp" && mv -f "$l_install_replace_temp" "$2"; then
        l_install_replace_temp=
        return 0
    fi
    rm -f "$l_install_replace_temp"
    l_install_replace_temp=
    return 1
}

l_install_cleanup() {
    l_install_status=$?
    trap - 0
    if [ "$l_install_done" != yes ]; then
        l_install_i=$l_install_count
        while [ "$l_install_i" -gt 0 ]; do
            l_install_entry=$l_install_work/state/$l_install_i
            if [ -f "$l_install_entry/committed" ]; then
                l_install_target=$(cat "$l_install_entry/target")
                if [ -f "$l_install_target" ] && [ ! -L "$l_install_target" ] && cmp -s "$l_install_target" "$l_install_entry/after" &&
                    [ "$(l_install_mode "$l_install_target")" = "$(l_install_mode "$l_install_entry/after")" ]; then
                    if [ -f "$l_install_entry/existed" ]; then
                        l_install_replace "$l_install_entry/before" "$l_install_target" ||
                            printf 'Could not restore %s; backup: %s\n' "$l_install_target" "$l_install_entry/before" >&2
                    else
                        rm -f "$l_install_target" || printf 'Could not remove %s\n' "$l_install_target" >&2
                    fi
                else
                    printf 'Preserved a concurrent edit to %s; original backup: %s\n' "$l_install_target" "$l_install_entry/before" >&2
                fi
            fi
            l_install_i=$((l_install_i - 1))
        done
        if [ "$l_install_binary_ready" = yes ]; then
            printf 'L is installed at %s/l, but terminal setup did not complete. Rerun the installer to repair setup.\n' "$l_install_prefix" >&2
        fi
        if [ -f "$l_install_work/install.log" ]; then
            printf 'Setup details and configuration backups: %s\n' "$l_install_work" >&2
        else
            rm -rf "$l_install_work"
        fi
    else
        rm -rf "$l_install_work" || :
    fi
    exit "$l_install_status"
}

l_install_download() {
    curl --proto '=https' --proto-redir '=https' --tlsv1.2 --fail --silent --show-error --location --retry 3 --connect-timeout 15 --max-time 120 "$@"
}

l_install_validate_version() {
    printf '%s\n' "$1" | awk 'NR == 1 && /^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[A-Za-z0-9]+([.-][A-Za-z0-9]+)*)?$/ { ok=1 } END { exit !(ok && NR == 1) }' ||
        l_install_fail 'Invalid release version.'
}

l_install_main() {
    set -eu
    export LC_ALL=C
    umask 077
    l_install_version=latest
    l_install_prefix=
    l_install_yes=no
    l_install_modify=yes
    while [ "$#" -gt 0 ]; do
        case $1 in
            --version|--prefix)
                [ "$#" -ge 2 ] && [ -n "$2" ] || l_install_fail "Missing value for $1"
                case $1 in --version) l_install_version=${2#v} ;; --prefix) l_install_prefix=$2 ;; esac
                shift 2 ;;
            --yes) l_install_yes=yes; shift ;;
            --no-modify-path) l_install_modify=no; shift ;;
            --help|-h)
                cat <<'L_HELP'
Usage: sh install.sh [--version VERSION] [--prefix DIRECTORY] [--yes] [--no-modify-path]

Install the latest stable L release and its Rust build tools on Apple Silicon
macOS or Linux ARM64/x86-64. The default executable directory is ~/.local/bin.
One confirmation permits tool downloads and Bash/Zsh startup-file changes.
--yes permits unattended setup; --no-modify-path leaves startup files unchanged.
The activation file is always written to ~/.local/share/l/env.
Native C/linker tools must already be installed. No sudo or system package manager is used.
L_HELP
                return ;;
            *) l_install_fail "Unknown argument: $1 (use --help)" ;;
        esac
    done
    : "${HOME:?HOME must identify your user directory}"
    [ -n "$l_install_prefix" ] || l_install_prefix=$HOME/.local/bin
    if [ "$l_install_version" != latest ]; then l_install_validate_version "$l_install_version"; fi
    for l_install_tool in bash sh curl uname cc awk sed head tail cat cmp cp mv rm mkdir mktemp dirname readlink tar sort ls; do
        if ! command -v "$l_install_tool" >/dev/null 2>&1; then
            if [ "$l_install_tool" = cc ]; then
                case $(uname -s) in
                    Darwin) l_install_fail 'Install Xcode Command Line Tools with xcode-select --install, then retry.' ;;
                    *) l_install_fail "Install your distribution's C compiler and linker development tools, then retry." ;;
                esac
            fi
            l_install_fail "Required command is missing: $l_install_tool"
        fi
    done
    case $(uname -s)/$(uname -m) in
        Darwin/arm64|Linux/aarch64|Linux/arm64|Linux/x86_64|Linux/amd64) ;;
        *) l_install_fail 'Supported systems are Apple Silicon macOS and Linux ARM64/x86-64. Intel macOS is not supported.' ;;
    esac
    if command -v sha256sum >/dev/null 2>&1; then
        l_install_digest() { sha256sum "$1" | awk '{print $1}'; }
    elif command -v shasum >/dev/null 2>&1; then
        l_install_digest() { shasum -a 256 "$1" | awk '{print $1}'; }
    else
        l_install_fail 'SHA-256 verification requires sha256sum or shasum.'
    fi
    l_install_prefix=$(l_install_resolve "$l_install_prefix") || exit 1
    l_install_cargo=$(l_install_resolve "${CARGO_HOME:-$HOME/.cargo}") || exit 1
    l_install_rustup=$(l_install_resolve "${RUSTUP_HOME:-$HOME/.rustup}") || exit 1
    l_install_env=$HOME/.local/share/l/env
    l_install_writable "$l_install_prefix/l"
    [ ! -L "$l_install_prefix/l" ] || l_install_fail 'The installation destination l is a symbolic link; choose another --prefix.'
    l_install_work=$(mktemp -d "${TMPDIR:-/tmp}/l-bootstrap.XXXXXXXX")
    l_install_count=0
    l_install_done=no
    l_install_binary_ready=no
    trap l_install_cleanup 0
    trap 'exit 130' INT
    trap 'exit 143' TERM
    mkdir "$l_install_work/state"
    printf 'int main(void) { return 0; }\n' > "$l_install_work/link-check.c"
    if ! cc "$l_install_work/link-check.c" -o "$l_install_work/link-check" > "$l_install_work/native-tools.log" 2>&1; then
        cat "$l_install_work/native-tools.log" >&2
        case $(uname -s) in
            Darwin) l_install_fail 'Native C/linker tools are unavailable. Run xcode-select --install and finish setup, then retry.' ;;
            *) l_install_fail "Install your distribution's C compiler and linker development tools, then retry." ;;
        esac
    fi
    l_install_stage "$l_install_env" env
    if [ "$l_install_modify" = yes ]; then
        l_install_stage "$HOME/.bashrc" profile
        l_install_login=$HOME/.profile
        for l_install_candidate in "$HOME/.bash_profile" "$HOME/.bash_login" "$HOME/.profile"; do
            if [ -e "$l_install_candidate" ] || [ -L "$l_install_candidate" ]; then l_install_login=$l_install_candidate; break; fi
        done
        l_install_stage "$l_install_login" profile
        l_install_zdot=${ZDOTDIR:-$HOME}
        l_install_stage "$l_install_zdot/.zshrc" profile
        l_install_stage "$l_install_zdot/.zprofile" profile
    fi
    if [ "$l_install_version" = latest ]; then
        l_install_latest=$(l_install_download --head --output /dev/null --write-out '%{url_effective}' https://github.com/l-programming-language/l/releases/latest) ||
            l_install_fail 'Could not resolve the latest release.'
        case $l_install_latest in
            https://github.com/l-programming-language/l/releases/tag/v*) l_install_version=${l_install_latest#https://github.com/l-programming-language/l/releases/tag/v} ;;
            *) l_install_fail 'Latest release did not resolve to a versioned GitHub release.' ;;
        esac
    fi
    l_install_validate_version "$l_install_version"
    printf 'Install L %s to %s\nRust tools: %s\n' "$l_install_version" "$l_install_prefix" "$l_install_cargo"
    if [ "$l_install_modify" = yes ]; then
        printf 'Configure Bash/Zsh startup files:\n'
        l_install_i=2
        while [ "$l_install_i" -le "$l_install_count" ]; do
            sed 's/^/  /' "$l_install_work/state/$l_install_i/logical"
            l_install_i=$((l_install_i + 1))
        done
    else
        printf 'Leave shell startup files unchanged.\n'
    fi
    if [ "$l_install_yes" != yes ]; then
        if ! ( : < /dev/tty ) 2>/dev/null; then l_install_fail 'No terminal is available for confirmation. Use --yes for unattended setup.'; fi
        printf 'Continue? [Y/n] ' > /dev/tty
        IFS= read -r l_install_reply < /dev/tty || l_install_fail 'Confirmation was not received.'
        case $l_install_reply in ''|y|Y|yes|YES) ;; *) l_install_fail 'Installation cancelled; no installation or shell changes were made.' ;; esac
    fi
    printf 'Downloading and verifying the release installer...\n'
    l_install_url=https://github.com/l-programming-language/l/releases/download/v$l_install_version
    l_install_download --output "$l_install_work/checksums.txt" "$l_install_url/checksums.txt" ||
        l_install_fail 'Could not download release checksums.'
    l_install_download --output "$l_install_work/install-release.sh" "$l_install_url/install-release.sh" ||
        l_install_fail 'Could not download the release installer.'
    l_install_expected=$(awk '$2 == "install-release.sh" { count++; hash=$1; if (NF != 2) bad=1 } END { if (count != 1 || bad) exit 1; print hash }' "$l_install_work/checksums.txt") ||
        l_install_fail 'Checksum inventory has a missing, duplicate, or invalid installer entry.'
    case $l_install_expected in ''|*[!a-f0-9]*) l_install_fail 'Invalid installer checksum.' ;; esac
    [ "${#l_install_expected}" -eq 64 ] || l_install_fail 'Invalid installer checksum.'
    [ "$(l_install_digest "$l_install_work/install-release.sh")" = "$l_install_expected" ] ||
        l_install_fail 'Release installer checksum does not match.'
    printf 'Installing L and preparing its Rust build tools...\n'
    if ! CARGO_HOME="$l_install_cargo" RUSTUP_HOME="$l_install_rustup" bash "$l_install_work/install-release.sh" --version "$l_install_version" --prefix "$l_install_prefix" --with-toolchain > "$l_install_work/install.log" 2>&1; then
        cat "$l_install_work/install.log" >&2
        l_install_fail 'The release installer did not complete.'
    fi
    l_install_binary_ready=yes
    [ "$("$l_install_prefix/l" --version)" = "$l_install_version" ] || l_install_fail 'Installed L reports an unexpected version.'
    l_install_i=1
    while [ "$l_install_i" -le "$l_install_count" ]; do
        l_install_unchanged "$l_install_work/state/$l_install_i" || exit 1
        l_install_i=$((l_install_i + 1))
    done
    printf 'Configuring terminal activation...\n'
    l_install_i=1
    while [ "$l_install_i" -le "$l_install_count" ]; do
        l_install_entry=$l_install_work/state/$l_install_i
        l_install_target=$(cat "$l_install_entry/target")
        l_install_unchanged "$l_install_entry" || exit 1
        mkdir -p "$(dirname "$l_install_target")"
        if ! cmp -s "$l_install_entry/before" "$l_install_entry/after" || [ ! -f "$l_install_entry/existed" ]; then
            l_install_replace "$l_install_entry/after" "$l_install_target" || l_install_fail "Could not update $l_install_target"
            : > "$l_install_entry/committed"
        fi
        l_install_i=$((l_install_i + 1))
    done
    l_install_checked=$(sh -c '. "$1"; command -v l; l --version' sh "$l_install_env") ||
        l_install_fail 'The installed environment could not activate L.'
    [ "$l_install_checked" = "$l_install_prefix/l
$l_install_version" ] || l_install_fail 'Terminal activation selected an unexpected L executable.'
    l_install_log=$(mktemp "$(dirname "$l_install_env")/install.XXXXXXXX")
    cp "$l_install_work/install.log" "$l_install_log"
    l_install_done=yes
    printf 'L %s is ready.\n' "$l_install_version"
    printf 'Activate this terminal: . "$HOME/.local/share/l/env"\nThen run: l --version\n'
    if [ "$l_install_modify" = yes ]; then printf 'New Bash/Zsh terminals will activate L automatically.\n'; fi
}

l_install_main "$@"
