<?php
namespace app\api\controller;

class Tong {
    
    //统计PV 页面访问量
    public function tjPV(){
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
    /*
     * 统计UV 用户访问量
     * 通过session识别
     */
    public function tjUV(){
        if(！empty($_COOKIE["access"]) && $_COOKIE["access"]==1){
            if(!file_exists("countu.txt")){
                $one_file=fopen("countu.txt","w+");
                
                fwrite("count.txtu","1");
                fclose("$one_file");
                setcookie("access",1, time()+3600*24); //访问过标记
            }else{
                $num=file_get_contents("countu.txt");
                $num++;
                file_put_contents("countu.txt","$num");
                $newnum=file_get_contents("countu.txt");
                setcookie("access",1, time()+3600*24);//访问过标记
            }
            return $newnum;
      }
    }
}