_minify_complete()
{
    local cur_work flags

    cur_word="${COMP_WORDS[COMP_CWORD]}"
    flags="-a --all -l --list --match --mime -o --output -r --recursive --type --url -v --verbose --version -w --watch --css-decimals --html-keep-conditional-comments --html-keep-default-attrvals --html-keep-document-tags --html-keep-end-tags --html-keep-whitespace --svg-decimals --xml-keep-whitespace"

    if [[ ${cur_word} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${flags}" -- ${cur_word}) )
    else
        COMPREPLY=()
    fi
    return 0
}

complete -F _minify_complete minify
