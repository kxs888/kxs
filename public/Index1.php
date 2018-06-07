<?php
namespace app\api\controller;

use app\common\controller\Common;

class Index extends Common
{
  
    public function login(){
        $res = array();
        $username = $this->request->post("username");
        $res['code'] = 0;
        $res['msg'] = '登录成功';
        
        return json_encode($res, JSON_UNESCAPED_UNICODE);
    }
    public function register(){
        $arr =array();
        $arr['phone'] = $this->request->post('phone');
        $result = Db::name('phone');
        echo $result;
        if(is_inarray($arr['phone'],$result)){
            
        }
        
        
        
    }
}