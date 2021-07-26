docker-compose -f docker-compose.dev.yml up --build && docker logs -f --since 5s msmf_dev
# docker-compose -f docker-compose.dev.yml up --build -d && docker attach --sig-proxy=false msmf_dev
