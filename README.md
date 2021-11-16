
# GIỚI THIỆU
Combo docker-compose cho Magento với các tính năng như:

* Chạy được đồng thời nhiều phiên bản PHP.
* Tự động tạo Virtual host cho nginx, hỗ trợ Magento 1, Magento 2, Laravel, Wordpress.
* Hỗ trợ SSL
* Tự động tải, cài đặt Magento fresh các phiên bản theo yêu cầu.
* Tạo/Drop/Import/Export database từ command.
* Bật tắt Xdebug ứng với từng phiên bản PHP.
* Email catch all local, tránh tình trạng gửi email ra ngoài internet (tất nhiên thì vẫn phải nhớ check xem config SMTP đừng có để configure của prod =)))

Hiện tại mới test trên Ubuntu, các hệ điều hành khác mọi người vui lòng tự mò =).

# HƯỚNG DẪN SỬ DỤNG
## Yêu cầu hệ thống
* Hệ thống cần có cài đặt docker và docker-compose. Hướng dẫn cài đặt có thể tham khảo Google hoặc:

[Hướng dẫn cài đặt docker trên Ubuntu](https://docs.docker.com/engine/install/ubuntu/)

[Hướng dẫn cài đặt docker-compose trên Ubuntu](https://docs.docker.com/compose/install/)

* Cài xong docker thì nhơ chạy lệnh này để add user máy mình đang chạy được quyền chạy docker mà không cần gõ sudo, gõ xong thì nhớ logout rồi login lại:
```bash
sudo usermod -aG docker $USER
```
### Giới thiệu hệ thống
```bash
> tree                                                                                                                                                   docker-magento -> master
.
├── README.md
├── build
│   ├── mailhog
│   │   └── Dockerfile
│   ├── nginx
│   │   ├── Dockerfile
│   │   └── conf
│   │       └── nginx.conf
│   ├── php70
│   │   └── Dockerfile
│   ├── php71
│   │   └── Dockerfile
│   ├── php72
│   │   └── Dockerfile
│   ├── php73
│   │   └── Dockerfile
│   └── php74
│       └── Dockerfile
├── conf
│   ├── nginx
│   │   ├── conf.d
│   │   ├── nginx.conf
│   │   └── ssl
│   └── php
│       ├── php70
│       │   ├── 10-opcache.ini
│       │   ├── magento.conf
│       │   └── php.ini
│       ├── php71
│       │   ├── 10-opcache.ini
│       │   ├── magento.conf
│       │   └── php.ini
│       ├── php72
│       │   ├── 10-opcache.ini
│       │   ├── magento.conf
│       │   └── php.ini
│       ├── php73
│       │   ├── 10-opcache.ini
│       │   ├── magento.conf
│       │   └── php.ini
│       └── php74
│           ├── 10-opcache.ini
│           ├── magento.conf
│           └── php.ini
├── data
├── databases
│   ├── export
│   └── import
├── docker-compose.yml
├── env-example
├── images
│   ├── cert.png
│   ├── cert02.png
│   └── cert03.png
├── logs
│   └── nginx
├── scripts
│   ├── create-vhost
│   ├── database
│   ├── fixowner
│   ├── init-magento
│   ├── list-services
│   ├── mysql
│   ├── setup-composer
│   ├── shell
│   ├── ssl
│   └── xdebug
└── sources
```
#### Cấu trúc thư mục hệ thống

| Thư mục | Chức năng |
|---------|-----------|
| build   | Chứa các file sử dụng trong quá trình build container sử dụng cho hệ thống |
| conf | Chứa các file config cho container sử dụng trong quá trình người dùng sử dụng |
| data | Chứa các dữ liệu cho các container như mysql, rabbitMQ |
| database | Folder sử dụng cho các chức năng import/export database |
| images | Folder ảnh của cái README.md này, LOL |
| logs | Folder chứa log nginx cho tiện theo dõi ngoài hệ thống |
| scripts | Folder chứa các command chức năng sử dụng cho hệ thống |
| sources | Folder chứa các thư mục sources của các website dự án |

#### Các services được cấu hình sẵn trong hệ thống
* Hệ thống đã được cấu hình sẵn các services sau:

| Tên services | Giải thích |
|--------------|------------|
| nginx | service webserver nginx |
| php70 | service php version php 7.0 |
| php71 | service php version php 7.1 |
| php72 | service php version php 7.2 |
| php73 | service php version php 7.3 |
| php74 | service php version php 7.4 |
| mysql | service mysql, default sử dụng version 8.0 |
| mailhog | service email catch all |
| elasticsearch | service Elastiscsearch |
| kibana | service Kibana |
| redis | service Redis  |
| rabbitMQ | service RabbitMQ  |

## Các lệnh docker/docker-compose cơ bản
* Xài docker thì cũng nên biết 1 số lệnh cơ bản sau:
```bash
# Xem thông tin các docker container đang chạy sử dụng docker-compose.yml tại thư mục hiện hành
docker-compose ps 
# Xem thông tin tài nguyên mà các container đang sử dụng
docker stats
# Khởi tạo toàn bộ các services (containers) được khai báo trong file docker-compose.yml
docker-compose up -d
# Stop và xoá toàn bộ các services (containers) tạo và đang chạy được khai vào trong file docker-compose.yml, bao gồm cả volumes (không bao gồm file trong thư mục ./data/)
docker-compose down --remove-orphans
# Tắt các services (container) đang chạy được khai báo trong file docker-compose.yml - Kiểu tắt 1 xíu cho đỡ nặng máy rồi tí khởi động lại.
docker-compose stop
# Khởi động các services (container) đã khởi tạo được khai báo trong file docker-compose.yml - services (container) nào mà không khởi tạo trước đó thì sẽ vẫn không được khởi tạo và không được start.
docker-compose start
# Restart lại các services (container) đang chạy
docker-composer restart
# Chui vô 1 services để chạy command - Ví dụ tính chui vô container service php72 để chạy composer
docker-compose exec php72 bash
```


# Test email with mailhog
```bash
/usr/local/bin/mhsendmail --smtp-addr="172.16.16.54:1025" test@mailhog.local <<EOF
From: Salman <kinsta@mailhog.local>
To: Test <test@mailhog.local>
Subject: Hello, MailHog!

Hey there,
Missing you pig time.

Hogs & Kisses,
Salman
EOF
```

# Install local SSL
## Install mkcert on Ubuntu
```bash
# Install mkcert
apt install libnss3-tools
wget https://github.com/FiloSottile/mkcert/releases/download/v1.4.3/mkcert-v1.4.3-linux-amd64 -O /usr/local/bin/mkcert
chmod +x /usr/local/bin/mkcert
# Generate Local CA
mkcert -install

# Create cert store folder
mkdir cd /path/to/project/cert
# Generate Local SSL Certificates
cd /path/to/project/cert
sudo mkcert example.com '*.example.com' localhost 127.0.0.1 ::1
# Cert will be store in the current folder 
```
![image info](./images/cert.png)

## Edit conf/443.conf with  the name of the cert file name, EG:
### From:
```bash
    ssl_certificate /etc/nginx/ssl/fullchain.pem;
    ssl_certificate_key /etc/nginx/ssl/privkey.pem;
```
### To:
```bash
    ssl_certificate /etc/nginx/cert/example.com+4.pem;
    ssl_certificate_key /etc/nginx/cert/example.com+4-key.pem;
```

## Uncomment mount docker-compose nginx-php services mount cert file
### From:
![image info](./images/cert02.png)

### To:
![image info](./images/cert03.png)

### Recreate nginx-php services
```bash
docker-compose up -d nginx-php
```

# EXTRA
## How to restart nginx or php on nginx-php container
Run this command inside nginx-php container
```bash
supervisorctl restart nginx
# Or
supervisorctl restart php-fpm
```
Or running this command inside docker-compose folder
```bash
docker-compose exec nginx-php supervisorctl restart nginx
# Or 
docker-compose exec nginx-php supervisorctl restart php-fpm
```