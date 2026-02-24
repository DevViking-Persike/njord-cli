# Njord CLI - Shell wrapper
# Replace the old njord alias with this function in your .zshrc:
#
# Remove or comment out: alias njord='source ~/.local/share/scripts/njord/njord.sh'
# Add the following:

njord() {
    local result
    result=$(~/.local/bin/njord-cli "$@" 2>/dev/tty)
    local code=$?
    if [[ $code -eq 0 && -n "$result" ]]; then
        eval "$result"
    fi
}
