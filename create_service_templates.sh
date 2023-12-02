#!/bin/bash

read -r -p "Path for configuration folder [~/.config/onvif-cam-poll]" $CONFIG_PATH
export CONFIG_PATH=${config_path:-~/.config/onvif-cam-poll}
echo "$CONFIG_PATH"
mkdir --p "$CONFIG_PATH"
export CONFIG_PATH
## create service templates
cp ./services/* "$CONFIG_PATH"/.
envsubst < "$CONFIG_PATH"/motion@.service > "$CONFIG_PATH"/motion@.service.temp && mv "$CONFIG_PATH"/motion@.service.temp "$CONFIG_PATH"/motion@.service
envsubst < "$CONFIG_PATH"/time@.service > "$CONFIG_PATH"/time@.service.temp && mv "$CONFIG_PATH"/time@.service.temp "$CONFIG_PATH"/time@.service
sudo cp "$CONFIG_PATH"/*.service /lib/systemd/system/.
sudo systemctl daemon-reload
echo "To track a new camera create an .env file for it in $CONFIG_PATH"
echo "Fill in all the variables that are listed in the example.env file"
echo "Afterwards, run systemctl start motion@camera_envfile_name"
echo "Don't forget to enable the service if you want it to run after a restart"