# Backend для мобильного приложения

Для запуска необходимо установить на машину nginx и docker

1. Добавить в `/etc/nginx/sites-available/seller-page.ddns.net.conf` файл [`seller-page.ddns.net.conf`](seller-page.ddns.net.conf):
   ```shell
    cp seller-page.ddns.net.conf /etc/nginx/sites-available/seller-page.ddns.net.conf
    
    sudo ln -s /etc/nginx/sites-available/seller-page.ddns.net.conf /etc/nginx/sites-enabled/seller-page.ddns.net.conf
    
    sudo nginx -t
    
    nginx -s reload
   ```

2. Запустить контейнер на порту 8082:
   ```shell
    docker build . -t seller-pages-image
    
    docker rm -f seller-pages-app 
   
    docker run --env-file ./.env --restart always -p 80:8080 -d --name seller-pages-app seller-pages-image:latest
   ```