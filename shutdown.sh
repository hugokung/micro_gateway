ps aux | grep mirco_gateway | grep -v 'grep' | awk '{print $2}' | xargs kill
