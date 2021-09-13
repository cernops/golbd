#/usr/bin/bash

#First, let's get the list of hostgroups
DEBUG=0

date
echo "Starting the report"

HOSTGROUPS=$(kermis -o read -a all -j | jq '.[] | ."hostgroup" ' | awk -F\" '{print $2}' | awk -F\/ '{print $1}' | sort | uniq -c | sort -n -r)
declare -A ALL_FE
while read -r line ;
do
  linearray=($line)
  FE=$(ai-pdb hostgroup_fact ${linearray[1]} fename --subgroups | jq '.[] | ."value" ' | sort |uniq -c |sort -n -r | head -n 1 | awk '{$1=""; print $0}')
  if [ "${FE}" == "" ] || [ "${FE}" == " \"Ignore\"" ] || [ "${FE}" == " \"ignore\"" ] || [ "${FE}" == " \"Unspecified\"" ];
  then
    echo "Error getting the FE of the hostgroup ${linearray[1]}"
    continue
  fi
  [ "$DEBUG" != 0 ] && echo " Hostgroup  ${linearray[1]} maps to FE $FE"
  if [ "${ALL_FE[$FE]}" != "" ];
  then 
    ALL_FE[$FE]=$(expr ${ALL_FE[$FE]} + ${linearray[0]} )
  else    
    ALL_FE[$FE]=${linearray[0]}
  fi
done <<< "$HOSTGROUPS"

ITEMS=""
for key in "${!ALL_FE[@]}"; do
  echo "The FE $key has ${ALL_FE[$key]}"
  [ "$ITEMS" != "" ] && ITEMS+=","
  ITEMS+="{\"ToChargeGroup\":$key, \"MetricValue\":${ALL_FE[$key]}}"
done
date=$(date -d yesterday +%Y-%m-%d)
DATA="""
{\"FromChargeGroup\":\"DNS Load Balancing\",
 \"MessageFormatVersion\":3,
 \"data\":[$ITEMS],
 \"TimeStamp\":\"$date\",
 \"TimeAggregate\":\"avg\",
 \"AccountingDoc\":\"http://configdocs.web.cern.ch/accounting\",
 \"TimePeriod\":\"day\",
 \"MetricName\":\"Number of aliases\"
 }
"""

API_KEY=$(tbag show lb_accounting_key --hg lxplus --plain)

#Finally, let's send the document
URL="https://accounting-receiver.cern.ch/v3/fe"
curl -X POST -H "Content-type: application/json" -H "API-Key: $API_KEY" $URL -d "$DATA"

echo "Report finished"
date
