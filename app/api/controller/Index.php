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
        
        $member = new Member();
        $member->phone = $this->request->post('phone');
        $member->username = $this->request->post('username'); 
        $member->passwd = md5($this->request->post('passwd')) ?? " ";
        $member->update_time = date('Y-m-d H:i:s', time());
        $member->create_time = date('Y-m-d H:i:s', time());
        $res = $member->save();
        if ($res) {
            $arr['code'] = 0;
            $arr['msg'] = '注册成功';
        } else {
            $arr['code'] = -1;
            $arr['msg'] = '注册失败';
        }
        
      
        return json_encode($arr);
    }
}