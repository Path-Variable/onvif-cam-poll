function check_var() {
  declare -n var_ref=$1
  declare -p "$var_ref" &>/dev/null
  declared=$?
  if ["$declared" != 0 ]
  then
    echo "$var_ref must be set! Exiting!"
    exit
  fi
}

for var in "$@"
do
  echo "checking $var is set"
  check_var "$var"
done


