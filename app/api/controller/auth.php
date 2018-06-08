<?php
namespace app\api\controller\auth;

use app\admin\common\tool\Weichat;

class Auth extends Controller
{
    // 验证用户是否微信登陆过
    public function initialize()
    {
        $user_id = session('user_id');
        if (! $user_id) {
            $path = $_SERVER['REQUEST_URI'];
            header('Location:/api/auth/index?path=' . $path);
        }
    }

    public function authCode()
    {
        $url = $this->request->post("url");
        $weichat = new Weichat($appId, $appSecret);
        $code = $weichat->getCode(urlEncode($url));
        return $code;
    }

    public function authUserInfo()
    {
        $code = $this->request->post['code'];
        $weichat = new Weichat($appId, $appSecret);
        $userinfo = $weichat->getUserInfo($code);
        if (empty($userinfo)) {
            $array['code'] = - 1;
            $array['msg'] = '授权失败';
        } else {
            $array['code'] = 0;
            $array['msg'] = '授权成功';
            
            $user = new User();
            $user['username'] = $userinfo['nickname'];
            $user['uid'] = $userinfo['openid'];
            $user['sex'] = $userinfo['sex'];
            $user_id = $user->where('uid', $usrinfo['openid'])->select();
            if (! $user_id) {
                $user->save();
            }
        }
        return $array;
    }
}