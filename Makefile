install:
	## check if vars are set
	bash ./check_vars.sh GOPATH
	## run go install
	go install ./...
	## create system-wide symlink to executables
	bash ./create_symlinks.sh motion-poll set-preset goto-preset set-time
	## create config folder
	bash ./create_service_templates.sh

