<?php
namespace app\api\controller;

class Tong {
    
    protected  static $num = 0;
    
    public function tj(){
       global $num; 
       $file_path = "E:/count.txt";
       if(!file_exists($file_path)){
           $one_file=fopen($file_path,"w+"); //建立一个统计文本，如果不存在就创建
           fwrite($file_path,"1");  //把数字1写入文本
           fclose($one_file);
       }else{ //如果不是第一次访问直接读取内容，并+1,写入更新后再显示新的访客数
           $num=file_get_contents($file_path);
           $num++;
           file_put_contents($file_path,"$num");
           $newnum=file_get_contents($file_path);
           fclose($file_path);
          return $newnum;
       }
        
    }
}