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
        
        return json_encode($res);
    }
}