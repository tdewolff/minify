if command -v dnf; then
    dnf -y install golang
elif command -v yum; then
    yum install epel-release
    yum -y install golang
elif command -v apk; then
    apk add go
fi
