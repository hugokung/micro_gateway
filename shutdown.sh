ps aux | grep go_gateway | grep -v 'grep' | awk '{print $2}' | xargs kill
