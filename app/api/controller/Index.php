<?php
namespace app\api\controller;

use app\common\controller\Common;
use app\api\model\Member;

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
        $member = new Member();
        $member->phone = $arr['phone'];
        $member->save();
      
        return json_encode($arr);
    }
}