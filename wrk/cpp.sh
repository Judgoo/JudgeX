wrk -t10 -c12 -d60s -s cpp.lua --timeout 2m http://127.0.0.1:8001/v1/judge/cpp
