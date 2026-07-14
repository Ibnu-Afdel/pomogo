complete -c pomogo -f
complete -c pomogo -n "not __fish_seen_subcommand_from version config stats history completion help status" -a "version config stats history completion help status"
complete -c pomogo -n "__fish_seen_subcommand_from config" -a "init"
complete -c pomogo -n "__fish_seen_subcommand_from stats" -l week -d "Show weekly activity"
complete -c pomogo -n "__fish_seen_subcommand_from stats" -l month -d "Show monthly summary"
complete -c pomogo -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
complete -c pomogo -n "__fish_seen_subcommand_from status" -l format -d "Output format (default, waybar, tmux, json)"
