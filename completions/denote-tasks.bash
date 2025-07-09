#!/bin/bash
# Bash completion for denote-tasks

_denote_tasks_completions() {
    local cur prev words cword
    _get_comp_words_by_ref -n : cur prev words cword

    local prog="${words[0]}"
    
    # Helper function to get task IDs
    _get_task_ids() {
        "$prog" completion task-ids 2>/dev/null
    }
    
    # Helper function to get project IDs/names
    _get_project_ids() {
        "$prog" completion project-ids 2>/dev/null | cut -d: -f1
    }
    
    # Helper function to get project names for display
    _get_project_names() {
        "$prog" completion project-ids 2>/dev/null
    }
    
    # Helper function to get areas
    _get_areas() {
        "$prog" completion areas 2>/dev/null
    }
    
    # Helper function to get tags
    _get_tags() {
        "$prog" completion tags 2>/dev/null
    }

    # Main command
    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "task project note --tui --help --version" -- "$cur"))
        return
    fi

    # Global flags available everywhere
    local global_flags="--config --dir --json --no-color --quiet -q --area --tui -t"

    # Check for entity type (task, project, note)
    local entity=""
    local subcmd=""
    
    for ((i=1; i<cword; i++)); do
        case "${words[i]}" in
            task|project|note)
                entity="${words[i]}"
                ;;
            new|list|update|done|edit|delete|log|tasks|rename)
                if [[ -n "$entity" ]]; then
                    subcmd="${words[i]}"
                fi
                ;;
        esac
    done

    # Task commands
    if [[ "$entity" == "task" ]]; then
        if [[ -z "$subcmd" ]]; then
            COMPREPLY=($(compgen -W "new list update done edit delete log $global_flags" -- "$cur"))
            return
        fi

        case "$subcmd" in
            new)
                case "$prev" in
                    -p|--priority)
                        COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                        ;;
                    -due|--due)
                        COMPREPLY=($(compgen -W "today tomorrow monday tuesday wednesday thursday friday saturday sunday" -- "$cur"))
                        ;;
                    -area|--area)
                        local areas=$(_get_areas)
                        COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                        ;;
                    -project|--project)
                        local projects=$(_get_project_ids)
                        COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                        ;;
                    -estimate|--estimate)
                        COMPREPLY=($(compgen -W "1 2 3 5 8 13 21" -- "$cur"))
                        ;;
                    *)
                        COMPREPLY=($(compgen -W "-p --priority -due --due -area --area -project --project -estimate --estimate -tags --tags $global_flags" -- "$cur"))
                        ;;
                esac
                ;;
                
            list)
                case "$prev" in
                    -status|--status)
                        COMPREPLY=($(compgen -W "open done paused delegated dropped" -- "$cur"))
                        ;;
                    -p|--priority)
                        COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                        ;;
                    -area|--area)
                        local areas=$(_get_areas)
                        COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                        ;;
                    -project|--project)
                        local projects=$(_get_project_ids)
                        COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                        ;;
                    -sort|--sort|-s)
                        COMPREPLY=($(compgen -W "modified priority due created" -- "$cur"))
                        ;;
                    *)
                        COMPREPLY=($(compgen -W "-all -a -status --status -area --area -p --priority -project --project -overdue --overdue -soon --soon -sort --sort -s -reverse --reverse -r $global_flags" -- "$cur"))
                        ;;
                esac
                ;;
                
            update|done|edit|delete|log)
                # These commands take task IDs as first argument
                if [[ "${words[cword-1]}" == "$subcmd" ]]; then
                    # Complete with task IDs
                    local task_ids=$(_get_task_ids)
                    COMPREPLY=($(compgen -W "$task_ids" -- "$cur"))
                elif [[ "$subcmd" == "update" ]]; then
                    case "$prev" in
                        -status|--status)
                            COMPREPLY=($(compgen -W "open done paused delegated dropped" -- "$cur"))
                            ;;
                        -p|--priority)
                            COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                            ;;
                        -area|--area)
                            local areas=$(_get_areas)
                            COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                            ;;
                        -project|--project)
                            local projects=$(_get_project_ids)
                            COMPREPLY=($(compgen -W "$projects" -- "$cur"))
                            ;;
                        *)
                            COMPREPLY=($(compgen -W "-status --status -p --priority -due --due -area --area -project --project -tags --tags $global_flags" -- "$cur"))
                            ;;
                    esac
                fi
                ;;
        esac
        
    # Project commands
    elif [[ "$entity" == "project" ]]; then
        if [[ -z "$subcmd" ]]; then
            COMPREPLY=($(compgen -W "new list update tasks $global_flags" -- "$cur"))
            return
        fi

        case "$subcmd" in
            new)
                case "$prev" in
                    -p|--priority)
                        COMPREPLY=($(compgen -W "p1 p2 p3" -- "$cur"))
                        ;;
                    -area|--area)
                        local areas=$(_get_areas)
                        COMPREPLY=($(compgen -W "$areas" -- "$cur"))
                        ;;
                    -status|--status)
                        COMPREPLY=($(compgen -W "active paused completed cancelled" -- "$cur"))
                        ;;
                    *)
                        COMPREPLY=($(compgen -W "-p --priority -area --area -status --status -due --due -tags --tags $global_flags" -- "$cur"))
                        ;;
                esac
                ;;
                
            list)
                case "$prev" in
                    -sort|--sort)
                        COMPREPLY=($(compgen -W "modified priority due created name area" -- "$cur"))
                        ;;
                    *)
                        COMPREPLY=($(compgen -W "-all -sort --sort -reverse --reverse $global_flags" -- "$cur"))
                        ;;
                esac
                ;;
        esac
        
    # Note commands
    elif [[ "$entity" == "note" ]]; then
        if [[ -z "$subcmd" ]]; then
            COMPREPLY=($(compgen -W "new list edit rename $global_flags" -- "$cur"))
            return
        fi

        case "$subcmd" in
            new)
                case "$prev" in
                    -tags|--tags)
                        local tags=$(_get_tags)
                        COMPREPLY=($(compgen -W "$tags" -- "$cur"))
                        ;;
                    *)
                        COMPREPLY=($(compgen -W "-tags --tags $global_flags" -- "$cur"))
                        ;;
                esac
                ;;
                
            list)
                case "$prev" in
                    -tag|--tag)
                        # Could complete with existing tags
                        ;;
                    *)
                        COMPREPLY=($(compgen -W "-tag --tag $global_flags" -- "$cur"))
                        ;;
                esac
                ;;
        esac
        
    # Legacy command support
    else
        case "${words[1]}" in
            add)
                # Same as task new
                _denote_tasks_completions
                ;;
            list)
                # Same as task list
                _denote_tasks_completions
                ;;
            done)
                # Complete with task IDs
                local task_ids=$(_get_task_ids)
                COMPREPLY=($(compgen -W "$task_ids" -- "$cur"))
                ;;
        esac
    fi
}

complete -F _denote_tasks_completions denote-tasks