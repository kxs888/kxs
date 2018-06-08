<?php
namespace app\api\controller;
use app\common\controller\Common;
use think\Db;
use app\api\model\User;
 use app\api\validate\Validate;


class Index extends Common
{
  
    public function login(){
        $array = array();
        $res = input('post.');
        $user = Db::name('user')->where('phone',$res['phone'])->find();
        if(empty($user)){
            $array['code'] = -1;
            $array['msg']  = '用户不存在';
            return json_encode($array,JSON_UNESCAPED_UNICODE);
        } 
        if (md5($res['passwd']) != ($user['passwd'])){
            $array['code'] = -1;
            $array['msg']  = '密码不正确';
        } else {
            $array['code'] = 0;
            $array['msg']  = '登陆成功';
        }
        return json_encode($array,JSON_UNESCAPED_UNICODE);
        
       
    }  
        
    public function register(){
        $array = array();
        $user1 = new User();
        $user1->username = input('username');
        $user1->phone = input('phone');
        $user1->passwd = md5(input('passwd'));
        $user1->create_time = date('Y-m-d H:i:s');
        $res = Db::name('user')->where('phone',$user1->phone)->find();
        if(!empty($res)){
            $array['code'] = -1;
            $array['msg'] = '已注册';
        } else {
        if (!Validate::isPhone($user1->phone)) {
            $array['code'] = -1;
            $array['msg'] = '手机号不正确';
        } elseif (!Validate::isUserName($user1->username)){
            $array['code'] = -1;
            $array['msg'] = '用户名格式不正确';
        } else {

            $res = $user1->save();
            if(!$res){
                $array['code'] = -1;
                $array['msg'] = '保存失败';
            } else  {
                $array['code'] = 0;
                $array['msg'] = '保存成功';
            } 
            
        }
//         return json_encode($array);
    }
        return json_encode($array);
    }   
}