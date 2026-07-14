_pomogo_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="version config stats history completion help status"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    case "${prev}" in
        config)
            COMPREPLY=( $(compgen -W "init" -- ${cur}) )
            return 0
            ;;
        stats)
            COMPREPLY=( $(compgen -W "--week --month" -- ${cur}) )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- ${cur}) )
            return 0
            ;;
        status)
            COMPREPLY=( $(compgen -W "--format" -- ${cur}) )
            return 0
            ;;
    esac
}
complete -F _pomogo_completion pomogo
