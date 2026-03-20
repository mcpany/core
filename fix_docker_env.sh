# Fix docker in docker by switching the storage driver to vfs which avoids overlayfs failures
cat << 'JSON' > /etc/docker/daemon.json
{
  "storage-driver": "vfs"
}
JSON

# Restart docker
kill -9 $(pgrep dockerd)
sleep 2
dockerd &
sleep 5
