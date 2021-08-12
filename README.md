# ip-blocker
Daemon running in k8s that can block ideally IPs reading from a datasource and writing to a cloud firewall


# Steps

1. Daemon every 10m read from ES logging-v2 and gets the lasts IPs banned via ratelimit
2. Order from highest to lowest ( times has been ratelimit or returned 429 ) 
3. For each IP looks for at cloud armor if the IP is already banned and if not, ban the IP
4. Every 1h there is another process inside the daemon that check if the expirationdate of the IP ban rules.
  If there is an expiration, it unban the IP
  Also set the blocked boolean field to true/false 
