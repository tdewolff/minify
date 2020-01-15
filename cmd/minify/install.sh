#!/bin/bash

# Adds the possibility to create a static executable.
# Must provide --static as first argument to enable.
if [[ ! ${CGO_ENABLED} == 0 && ${1} == "--static" ]]; then
	export CGO_ENABLED=0; shift;
fi

# Compile/strip/install with variable data injection
go install -ldflags "-s -w -X 'main.Version=devel' -X 'main.Commit=$(git rev-list -1 HEAD)' -X 'main.Date=$(date)'";

# Completion url, redirected from https://git.io
COMPURL=https://git.io/minify-completion.bash;

# Bash completion directory variables
BASH_COMPLETION_DIR=${1};

# Install bash completions to proper directory, if any found
if [[ -n "${BASH_COMPLETION_DIR}" ]]; then
	echo -e "User-defined directory: ${BASH_COMPLETION_DIR}\n" >&2;

elif [[ -n "${BASH_COMPLETION_DIR_COMPAT_DIR}" ]]; then
	# Added for MACOS support with Homebrew
	BASH_COMPLETION_DIR=${BASH_COMPLETION_COMPAT_DIR};

elif [[ -r /usr/local/etc/bash_completion.d ]]; then
	# Added for MACOS support with Homebrew
	BASH_COMPLETION_DIR=/usr/local/etc/bash_completion.d;

elif [[ -n "${BASH_COMPLETION_USER_DIR}" ]]; then
	BASH_COMPLETION_DIR=${BASH_COMPLETION_USER_DIR};

elif [[ -r /usr/share/bash-completion/completions ]]; then
	BASH_COMPLETION_DIR=/usr/share/bash-completion/completions;

elif [[ -r /etc/bash_completion.d ]]; then
	BASH_COMPLETION_DIR=/etc/bash_completion.d;

fi

if [[ -r ${BASH_COMPLETION_DIR} ]]; then
	echo "Installing bash completion into ${BASH_COMPLETION_DIR}";
	# Try installing from URL first, if that fails, resort to using the existing file
	curl -sLko ${BASH_COMPLETION_DIR}/minify.bash ${COMPURL} 2>/dev/null || [[ -f minify_bash_tab_completion ]] && cat minify_bash_tab_completion >${BASH_COMPLETION_DIR}/minify.bash 2>/dev/null;

else
	# Could not find directory, show error message
	echo -e "Could not find bash completion directory. Try running again\nwith a valid completion  directory as parameter or manually\ndownload  and  save:  ${COMPURL}" >&2;

fi

# Source file for immediate use
source minify_bash_tab_completion;
