function create_symlink() {
  sudo rm "/usr/bin/$1"
  sudo ln -s "$GOPATH/$1" "/usr/bin/$1"
}

for var in "$@"
do
  echo "creating symlink for $var"
  create_symlink "$var"
done