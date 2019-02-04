package templates

// ConfigMapNginx Nginx.conf file
var ConfigMapNginx = `
error_log /proc/self/fd/2;
pid /var/run/nginx.pid;
user root;
worker_processes auto;
worker_rlimit_nofile 500000;

events {
	multi_accept on;
	use epoll;
	worker_connections 8192;
}

http {
	access_log /proc/self/fd/1;
	client_max_body_size 20m;
	default_type application/octet-stream;
	gzip on;
	gzip_buffers 16 8k;
	gzip_comp_level 4;
	gzip_disable msie6;
	gzip_proxied off;
	gzip_types application/json;
	gzip_vary on;
	include /etc/nginx/mime.types;
	index index.html index.htm;
	keepalive_timeout 240;
	proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=one:8m max_size=3000m inactive=600m;
	proxy_temp_path /var/tmp;
	sendfile on;
	server_tokens off;
	tcp_nopush on;
	types_hash_max_size 2048;

	server {
			#IPv4
			listen 80;

			#IPv6
			listen [::]:80;

			# Filesystem root of the site and index with fallback.
			root /var/www/html;
			index index.php index.html index.htm;

			# Make site accessible from http://domain;
			server_name drupal.innovation.cloud.statcan.ca;

			location / {
					# First attempt to serve request as file, then
					# as directory, then fall back to displaying a 404.
					try_files $uri $uri/ /index.html /index.php?$query_string;
			}

			location ~ \.php$ {
				proxy_intercept_errors on;
				fastcgi_split_path_info ^(.+\.php)(/.+)$;
				#NOTE: You should have "cgi.fix_pathinfo = 0;" in php.ini
				include fastcgi_params;
				fastcgi_read_timeout 120;
				fastcgi_param SCRIPT_FILENAME $request_filename;
				fastcgi_intercept_errors on;
				fastcgi_pass [[ .Host ]]:9000;
			}

			location ~ (^/s3/files/styles/|^/sites/.*/files/imagecache/|^/sites/.*/files/styles/) {
				expires max;
				try_files $uri @rewrite;
			}

			location ~* ^/(s3fs-css|s3fs-js|sites/default/files)/(.*) {
				set $s3_base_path "stcdrupal.blob.core.windows.net/drupal-public";
				set $file_path $2;

				resolver 10.0.0.10 valid=5s ipv6=off;
				resolver_timeout 5s;

				proxy_pass https://$s3_base_path/$file_path;
			}

			location ~ /\.ht {
				deny all;
			}
	}
}
`
