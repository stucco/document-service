Vagrant.configure("2") do |config|

  config.vm.hostname = "stucco-doc-service"

  config.vm.box = "vagrant_ubuntu_12.04.3_amd64_virtualbox"
  config.vm.box_url = "http://nitron-vagrant.s3-website-us-east-1.amazonaws.com/vagrant_ubuntu_12.04.3_amd64_virtualbox.box"

  config.vm.synced_folder "./", "/stucco/document-service"

  config.vbguest.auto_update = false

  # Fix docker not being able to resolve private registry in VirtualBox
  config.vm.provider :virtualbox do |vb, override|
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    vb.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
  end

  config.vm.provision :shell, :inline => "echo 'Running apt-get update' ; sudo apt-get update"

  config.vm.provision :shell, :inline => "echo 'Installing curl' ; sudo apt-get install curl -y"

  config.vm.provision :shell, :inline => "echo 'Installing docker' ; curl -s https://get.docker.io/ubuntu/ | sudo sh"

  config.vm.provision :shell, :inline => "echo 'Configuring docker' ; sudo groupadd docker ; sudo gpasswd -a vagrant docker ; sudo service docker restart"

end