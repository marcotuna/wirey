# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

$discovery=<<SCRIPT
yum update -y
yum install curl -y
cd /tmp
curl -fsSL get.docker.com -o get-docker.sh
sh get-docker.sh
systemctl enable --now docker
docker rm -f etcd
docker run  --restart always --name etcd \
  -d --net=host quay.io/coreos/etcd:v3.3 \
  /usr/local/bin/etcd \
  --listen-client-urls http://192.168.33.10:2379 \
  --advertise-client-urls http://192.168.33.10:2379
docker run --restart always --name vault \
-d --net=host --cap-add=IPC_LOCK vault:1.0.1 vault server -dev \
-dev-listen-address=192.168.33.10:8200 \
-dev-root-token-id=toor
export VAULT_ADDR=http://192.168.33.10:8200
export VAULT_TOKEN=toor
SCRIPT
$bootstrap=<<SCRIPT
yum update -y
yum install libmnl-devel gcc make kernel-devel wget -y
cd /tmp

wget https://git.zx2c4.com/WireGuard/snapshot/WireGuard-0.0.20181218.tar.xz
tar -xvf WireGuard-0.0.20181218.tar.xz
cd WireGuard-0.0.20181218/src/
make
make install
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  num_nodes = 3
  base_ip = "192.168.33."
  net_ips = num_nodes.times.collect { |n| base_ip + "#{n+11}" }

  config.vm.define "discovery-server" do |discovery|
    discovery.vm.box = "centos/7"
    discovery.vm.hostname = "discovery-server"
    discovery.vm.network :private_network, ip: "192.168.33.10"
    discovery.vm.provider "virtualbox" do |vb|
     vb.customize ["modifyvm", :id, "--memory", "512"]
    end
    discovery.vm.provision :shell, inline: $discovery
  end

  num_nodes.times do |n|
    config.vm.define "net-#{n+1}" do |net|
      net.vm.box = "centos/7"
      net_ip = net_ips[n]
      net_index = n+1
      net.vm.hostname = "wirey-node-#{net_index}"
      net.vm.provider "virtualbox" do |vb|
        vb.customize ["modifyvm", :id, "--memory", "1024"]
      end
      net.vm.network :private_network, ip: "#{net_ip}"
      net.vm.provision :shell, inline: $bootstrap, :args => "#{net_ip}"
    end
  end
end
