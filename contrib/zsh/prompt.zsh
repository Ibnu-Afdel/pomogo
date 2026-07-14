# Zsh prompt snippet for PomoGo
# Add this to your ~/.zshrc

# Helper to fetch active PomoGo session info
function pomogo_prompt_info() {
    if (( $+commands[pomogo] )); then
        local status_out
        status_out=$(pomogo status --format waybar 2>/dev/null)
        if [[ -n "$status_out" ]]; then
            if (( $+commands[jq] )); then
                local text
                text=$(echo "$status_out" | jq -r '.text' 2>/dev/null)
                if [[ -n "$text" ]]; then
                    echo " %F{red}🍅 $text%f"
                fi
            fi
        fi
    fi
}

# Example integration with PS1/PROMPT:
# PROMPT='$(pomogo_prompt_info) '$PROMPT
