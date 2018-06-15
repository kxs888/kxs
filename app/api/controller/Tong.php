<?php
namespace app\api\controller;

class Tong {
    
    protected  static $num = 0;
    
    public function tj(){
       global $num; 
     
       if(!file_exists("count.txt")){
           $one_file=fopen("count.txt","w+"); //建立一个统计文本，如果不存在就创建
           fwrite("count.txt","1");  //把数字1写入文本
           fclose($one_file);
       }else{ //如果不是第一次访问直接读取内容，并+1,写入更新后再显示新的访客数
           $num=file_get_contents("count.txt");
           $num++;
           file_put_contents("count.txt","$num");
           $newnum=file_get_contents("count.txt");
//            fclose("count.txt");
          return $newnum;
       }
        
    }
}