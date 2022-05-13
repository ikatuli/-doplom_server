# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Базовый образ
  # config.vm.box = "archlinux/archlinux"
  config.vm.box = "./images/Arch-Linux-x86_64-virtualbox-20220418.0.box"

  # Название хоста
  config.vm.hostname = "Doplom"
  config.vm.define "Doplom"

  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  # config.vm.box_check_update = false

  # Открываем порт 9000 только для локальных запросов
  config.vm.network "forwarded_port", guest: 3128, host: 3128, host_ip: "127.0.0.1"
  config.vm.network "forwarded_port", guest: 9000, host: 9000, host_ip: "127.0.0.1"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  # config.vm.synced_folder "../data", "/vagrant_data"

  # Выдаём больше памяти
  config.vm.provider "virtualbox" do |vb|
     vb.memory = "2024"
  end
  #
  # View the documentation for the provider you are using for more
  # information on available options.

  # Обновление системы
   config.vm.provision "shell", inline: <<-SHELL
      pacman -Suy gcc-go postgresql squid openssl git base-devel clamav --noconfirm
      yes | pacman -Scc
   SHELL

   # Синхронизация каталога репозитория с виртуальной машиной
   config.vm.synced_folder ".", "/app"
   #Сборка прогарммы
   config.vm.provision "shell", inline: <<-SHELL
     cd /app/
     go mod download
     go mod tidy
     go build -compiler=gccgo main.go
   SHELL

   # Инициализация базы данных
   config.vm.provision "shell", inline: <<-SHELL
     sudo -u postgres initdb -E UTF8 -D /var/lib/postgres/data
     systemctl start postgresql.service 
     sudo -u postgres psql --command "CREATE USER test_user WITH PASSWORD 'test';"
     sudo -u postgres createdb -O test_user test_bd
     systemctl enable postgresql.service --now
   SHELL

   # Создадим юнит сервера
   config.vm.provision "shell", inline: <<-SHELL
     echo "[Unit]
Description=Управляющий сервер doplom
[Service]
WorkingDirectory=/app
Type=simple
ExecStart=/app/main
[Install]
WantedBy=multi-user.target" > /etc/systemd/system/doplom.service
     systemctl daemon-reload
     systemctl enable doplom.service --now
   SHELL

   config.vm.provision "shell", inline: "systemctl enable squid.service --now"

   #Собираем e2guardian
   config.vm.provision "shell", inline: <<-SHELL
    mkdir /tmp/e2guardian
    cd /tmp/e2guardian/
    echo "[Unit]
Description=E2guardian web filtering
After=network.target

[Service]
Type=forking
ExecStart=/usr/bin/e2guardian

[Install]
WantedBy=multi-user.target" > /tmp/e2guardian/e2guardian.service

    echo "post_install() {
chown -R nobody:nobody /var/log/e2guardian
}" > /tmp/e2guardian/e2guardian.install


   cp /app/e2guardian_PKGBUILD /tmp/e2guardian/PKGBUILD
   chown vagrant:vagrant -R /tmp/e2guardian/
   sudo -u vagrant makepkg 
   pacman -U e2guardian-v5.4.3r* --noconfirm
   systemctl status e2guardian
   SHELL

   # Делаем копии конфигурации
   config.vm.provision "shell", inline: <<-SHELL
     cp /etc/e2guardian/e2guardian.conf /etc/e2guardian/e2guardian.conf.old
     cp /etc/squid/squid.conf /etc/squid/squid.conf.old
     cp /etc/clamav/freshclam.conf /etc/clamav/freshclam.conf.old
     cp /etc/e2guardian/contentscanners/clamdscan.conf /etc/e2guardian/contentscanners/clamdscan.conf.old
   SHELL

   config.vm.provision "shell", inline: <<-SHELL
    sed -i 's|database.clamav.net|https://packages.microsoft.com/clamav/|g' /etc/clamav/freshclam.conf
    freshclam
   SHELL
end
