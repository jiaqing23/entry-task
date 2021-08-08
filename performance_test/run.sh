echo "" >> wrk.txt
wrk -t200 -c1000 -d10s -s login.lua --timeout 20 http://localhost:3000/ >> wrk.txt &


for i in {1..10}; do 
#echo "%CPU %MEM ARGS $(date)" && ps -e -o pid,pcpu,pmem,args --sort=pcpu
echo "------------Memory & CPU Usage------------" >> log.txt
top -b -n 1 >> log.txt
echo "" >> log.txt
#sudo lsof | grep TCP  >> log.txt; 
echo "------------Network Connection------------" >> log.txt
sudo find /proc -maxdepth 1 -type d -name '[0-9]*' \
     -exec bash -c "ls {}/fd/ | wc -l | tr '\n' ' '" \; \
     -printf "fds (PID = %P), command: " \
     -exec bash -c "tr '\0' ' ' < {}/cmdline" \; \
     -exec echo \; | sort -rn | head >> log.txt
echo "" >> log.txt

echo "--------------IO Utilization--------------" >> log.txt
pidstat -dl >> log.txt
echo "" >> log.txt
sleep 1
done &
