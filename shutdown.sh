ps aux | grep micro_gateway | grep -v 'grep' | awk '{print $2}' | xargs kill
