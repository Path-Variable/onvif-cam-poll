install:
	## check if vars are set
	bash ./install_scripts/check_vars.sh GOPATH
	## run go install
	go install ./...
	## create system-wide symlink to executables
	bash ./install_scripts/create_symlinks.sh onvif-motion-poll onvif-set-preset onvif-goto-preset onvif-set-time
	## create config folder
	bash ./install_scripts/create_service_templates.sh

