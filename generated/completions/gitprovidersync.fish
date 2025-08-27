# gitprovidersync fish shell completion

function __fish_gitprovidersync_no_subcommand --description 'Test if there has been any subcommand yet'
    for i in (commandline -opc)
        if contains -- $i sync status print man help h completion
            return 1
        end
    end
    return 0
end

complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l config-file -r -d 'Path to the configuration file'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l quiet -d 'Equivalent to --verbosity=quiet. Only output errors'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l verbosity -r -d 'Set output verbosity: quiet | brief | verbose | debug | trace'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l output-format -r -d 'Output format (console,json,plain)'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l config-file-only -d 'Read configuration from file only (ignore ENV, dotenv, XDG_CONFIG_HOME)'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l plain -d 'Equivalent to --output-format=plain. Tabular text output for pipelines'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l verbosity-with-caller -d 'Add caller information to verbosity output (for development)'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l help -s h -d 'show help'
complete -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -f -l version -s v -d 'print the version'
complete -x -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -a 'sync' -d 'Mirror repositories from a source Git provider to targets'
complete -c gitprovidersync -n '__fish_seen_subcommand_from sync' -f -l dry-run -d 'Show what would be synced without making changes'
complete -c gitprovidersync -n '__fish_seen_subcommand_from sync' -f -l force-push -d 'Force push to target repositories'
complete -c gitprovidersync -n '__fish_seen_subcommand_from sync' -f -l active-from-limit -r -d 'Only sync repositories active since this date/time'
complete -c gitprovidersync -n '__fish_seen_subcommand_from sync' -f -l alpha-num-hyph-name -d 'Clean repository names to alphanumeric + hyphens only'
complete -c gitprovidersync -n '__fish_seen_subcommand_from sync' -f -l ignore-invalid-name -d 'Ignore repositories with invalid names'
complete -c gitprovidersync -n '__fish_seen_subcommand_from sync' -f -l help -s h -d 'show help'
complete -x -c gitprovidersync -n '__fish_seen_subcommand_from sync; and not __fish_seen_subcommand_from help h' -a 'help' -d 'Shows a list of commands or help for one command'
complete -x -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -a 'status' -d 'Show system status and suggest next actions'
complete -c gitprovidersync -n '__fish_seen_subcommand_from status' -f -l connectivity-check -d 'Test connectivity to configured providers'
complete -c gitprovidersync -n '__fish_seen_subcommand_from status' -f -l skip-suggestions -d 'Don\'t show suggested next actions'
complete -c gitprovidersync -n '__fish_seen_subcommand_from status' -f -l help -s h -d 'show help'
complete -x -c gitprovidersync -n '__fish_seen_subcommand_from status; and not __fish_seen_subcommand_from help h' -a 'help' -d 'Shows a list of commands or help for one command'
complete -x -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -a 'print' -d 'Print the current configuration'
complete -c gitprovidersync -n '__fish_seen_subcommand_from print' -f -l connectivity-check -d 'Test connectivity to configured providers'
complete -c gitprovidersync -n '__fish_seen_subcommand_from print' -f -l help -s h -d 'show help'
complete -x -c gitprovidersync -n '__fish_seen_subcommand_from print; and not __fish_seen_subcommand_from help h' -a 'help' -d 'Shows a list of commands or help for one command'
complete -c gitprovidersync -n '__fish_seen_subcommand_from man' -f -l help -s h -d 'show help'
complete -x -c gitprovidersync -n '__fish_seen_subcommand_from man; and not __fish_seen_subcommand_from help h' -a 'help' -d 'Shows a list of commands or help for one command'
complete -x -c gitprovidersync -n '__fish_gitprovidersync_no_subcommand' -a 'help' -d 'Shows a list of commands or help for one command'
complete -c gitprovidersync -n '__fish_seen_subcommand_from completion' -f -l help -s h -d 'show help'
complete -x -c gitprovidersync -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from help h' -a 'help' -d 'Shows a list of commands or help for one command'
