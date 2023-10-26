#/bin/bash 

for configfile in ./configs/*.{yaml,conf}; do 
	filename=$(basename $configfile)
	filename="${filename%.*}"

	docker config rm "${filename}_config" 2>/dev/null
	docker config create "${filename}_config" $configfile
done

