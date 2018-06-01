<?php
namespace app\api\Validate;

class Validate {
    
    public function isPhone($phone){
        if(strlen($phone) == 11){
            $n = preg_match("/13[123569]{1}\d{8}|15[1235689]\d{8}|188\d{8}/",$phone);
            if($n != 0){
                return true;
            }
        } else{
            return false;
        }
        
    function isUsername($username){
        //中英文6-20个字符正则
        $preg='/^[a-zA-Z\x{4e00}-\x{9fa5}]{6,20}$/u';
        $n = preg_match($preg,$username);
        if($n != 0){
            return true;
        } else {
            return false;
        }
        
    function isPWD($passwd){
        $n = preg_match('/^([a-zA-Z0-9]){6,16}$/', $password);
        if($n != 0){
            return true;
        } else {
            return false;
        }
        
      }
     }
    }
   }
