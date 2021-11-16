
# GIỚI THIỆU
Combo docker-compose cho Magento với các tính năng như:

* Chạy được đồng thời nhiều phiên bản PHP.
* Tự động tạo Virtual host cho nginx, hỗ trợ Magento 1, Magento 2, Laravel, Wordpress.
* Hỗ trợ SSL
* Tự động tải, cài đặt Magento fresh các phiên bản theo yêu cầu.
* Tạo/Drop/Import/Export database từ command.
* Bật tắt Xdebug ứng với từng phiên bản PHP.
* Tự động add các domain sử dụng vào /etc/hosts.
* Email catch all local, tránh tình trạng gửi email ra ngoài internet (tất nhiên thì vẫn phải nhớ check xem config SMTP đừng có để configure của prod =)))

Hiện tại mới test trên Ubuntu, các hệ điều hành khác mọi người vui lòng tự mò =).

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

#### Các command của hệ thống
| Command | Tác dụng |
|---------|----------|
| create-vhost | Tự động tạo virtual host cho services nginx ứng với từng loại magento và version php |
| database | create/drop/import/export/list databases | 
| fixowner | Command sử dụng để change lại owner của thư mục source code đúng với default mà hệ thống sử dụng |
| init-magento | Command sử dụng để tự động tải về và cài đặt Magento lên hệ thống |
| list-services | Command sử dụng để list các services mà docker-compose đã khởi tạo và đang chạy |
| mysql | Command sử dụng để tương tác với mysql shell trong mysql container |
| setup-composer | Composer sử dụng để setup auth.json default cho repo của Magento trong trường hợp cần thiết |
| shell | Command sử dụng để truy cập vào các container php, nginx, mysql |
| ssl | Command sử dụng để tạo Virtual host SSL cho các domain được lựa chọn |
| xdebug | Command sử dụng để bật/tắt xdebug của 1 service php được lựa chọn |

# HƯỚNG DẪN SỬ DỤNG
## Các lệnh docker/docker-compose cơ bản
* Xài docker thì cũng nên biết 1 số lệnh cơ bản sau:
```bash
# Xem thông tin các docker container đang chạy sử dụng docker-compose.yml tại thư mục hiện hành
docker-compose ps 

# Xem thông tin tài nguyên mà các container đang sử dụng
docker stats

# Khởi tạo toàn bộ các services (containers) được khai báo trong file docker-compose.yml
docker-compose up -d

# Khởi tạo và chạy một số services (container) được lựa chọn, chứ không khởi động toàn bộ services (container) được khai báo trong docker-compose.yml - Ví dụ chỉ khởi tạo và chạy nginx, php72, mysql
docker-compose up -d nginx php72 mysql

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

## Hướng dẫn sử dụng hệ thống

* Clone repo này vào một thư mục trên máy
* Copy file env-example thành .env
* Đổi thông tin cần thiết nếu có trong file .env trước khi chạy.

**Lưu ý:**
* Mọi command khi chạy trên hệ thống cần chạy trong thư mục chứa file docker-compose.yml
* Source code website cần được để trong 1 thư mục riêng trong thư mục sources. Nên để tạo thư mục dạng sources/domain.com và clone sources code vào thư mục này. Source code nên để trực tiếp trong thư mục sources/domain.com/ chứ không để trong sources/domain.com/src, trong trường hợp để trong sources/domain.com/src cần lưu ý trong bước tạo virtual host của nginx.
* Các command sử dụng cần được gọi theo dạng ./scripts/ten_command. Ví dụ: ./scripts/xdebug enable --php-version=php72
* Các command đều có hướng dẫn sử dụng riêng, có thể xem hướng dẫn bằng cách gõ command trong shell, ví dụ:
```bash
user@local:~/docker-magento$./scripts/xdebug
Docker Xdebug tools
Version 1

./scripts/xdebug [OPT] [ARG]...

    Options:
        enable                    Enable Xdebug.
        disable                   Disable Xdebug.
        status                    List all Xdebug status.
    Args:
        --php-version             PHP version used for Xdebug (php70|php71|php72|php73|php74).
        -h, --help                Display this help and exit.

    Examples:
      Disable Xdebug for PHP 7.2
        ./scripts/xdebug disable --php-version=php72
      Enable Xdebug for PHP 7.3
        ./scripts/xdebug enable --php-version=php73


                ____  __  __    _    ____ _____ ___  ____   ____
               / ___||  \/  |  / \  |  _ \_   _/ _ \/ ___| / ___|
               \___ \| |\/| | / _ \ | |_) || || | | \___ \| |
                ___) | |  | |/ ___ \|  _ < | || |_| |___) | |___
               |____/|_|  |_/_/   \_\_| \_\|_| \___/|____/ \____|



################################################################################
```

## Một số ví dụ

### Khởi tạo và chạy nginx, php72, mysql

```bash
docker-compose up -d nginx php72 mysql
```
