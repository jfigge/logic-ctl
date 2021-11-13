#!/bin/bash
cp ../adc* .
mv $1_abs.xml $2_abs.xml
mv $1_abx.xml $2_abx.xml 
mv $1_aby.xml $2_aby.xml 
mv $1_imm.xml $2_imm.xml 
mv $1_izx.xml $2_izx.xml 
mv $1_izy.xml $2_izy.xml 
mv $1_zpg.xml $2_zpg.xml 
mv $1_zpx.xml $2_zpx.xml 
LC_ALL=C find . -type f -name '*.xml' -exec sed -i '' s/$1/$2/ {} + 
mv *.xml ..
