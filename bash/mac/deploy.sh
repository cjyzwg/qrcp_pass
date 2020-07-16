#!/usr/bin/env bash
if [ -f "static.tar.gz" ];then
  echo "文件存在"
else
  echo "文件不存在"
  command1 = `wget https://gitee.com/cjyzwg/qrcp_pass/raw/qrcp_static/static.tar.gz`
fi
if [ -d "./public/" ];then
  echo "文件夹已经存在"
else
  tar -zxvf static.tar.gz
fi

./main $1

