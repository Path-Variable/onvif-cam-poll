if pgrep -s 0 '^sudo$' > /dev/null ; then
    echo "Running with sudo privileges"
else
    echo "You must run this script with sudo privileges!"
    exit 1
fi