install:
	## check if vars are set
	source ./check_vars.sh GOPATH
	## check for sudo privileges
	source ./check_sudo.sh
	## run go install
	go install ./...
	## create system-wide symlink to executables
	source ./create_symlinks.sh motion-poll set-preset goto-preset set-time
	## create config folder
	echo "Path for configuration folder [~/.config/onvif-cam-poll]"
	read path
	if [ -z "$path" ]
	then path="~/.config/onvif-cam-poll"
	mkdir -p $path
	## create service templates
		cp ./services/* $path/.
        ## overwrite service files - environment file, user, group, execstart
        ## copy service files to system service folder
        ## run systemctl daemon-reload

