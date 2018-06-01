<?php
namespace app\api\controller;
use app\common\controller\Common;
use app\api\model\Member;
// use app\api\Validate;
// include '../validate/Validate.php';
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
     
        
        $arr = array();
        $mem = new Member();
        $mem->phone = input('username');
        $mem->username = input('username');
        $mem->passwd = input('passwd');
        if (!isPhone($men->phone)) {
            $arr['code'] = -1;
            $arr['msg'] = '电话不正确';
            
        } elseif (!isUsername($men->username)){
            
            $arr['msg'] = '用户名格式不正确';
            
        } elseif(!isPWD($men->passwd)){
            
            $arr['msg'] = '密码不正确';
        } else{
            $arr['code'] = 0;
            $arr['msg'] = '注册成功';
        }
        
              return json_encode($arr);
        
        
    }
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
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
        
    }
//         $arr =array();
        
//         $member = new Member();
//         $member->phone = $this->request->post('phone');
//         if(!$this->isPhone($member->phone)){
//             $arr['code'] = -1;
//             $arr['msg'] = '注册失败';
//             return $arr;
            
//         }
        
//         $member->username = $this->request->post('username'); 
//         $member->passwd = md5($this->request->post('passwd')) ?? " ";
// //         $member->update_time = date('Y-m-d H:i:s', time());
//         $member->create_time = time();
//         $res = $member->save();
//         if ($res) {
//             $arr['code'] = 0;
//             $arr['msg'] = '注册成功';
//         } else {
//             $arr['code'] = -1;
//             $arr['msg'] = '注册失败';
//         }
        
      
//         return json_encode($arr);
//     }
//     public function isPhone($tel){
//         //正则表达式
     
//         if(strlen($tel) == "11")
//         {
//             //上面部分判断长度是不是11位
//             $n = preg_match_all("/13[123569]{1}\d{8}|15[1235689]\d{8}|188\d{8}/",$tel,$array);
//             /*接下来的正则表达式("/131,132,133,135,136,139开头随后跟着任意的8为数字 '|'(或者的意思)
//              * 151,152,153,156,158.159开头的跟着任意的8为数字
//              * 或者是188开头的再跟着任意的8为数字,匹配其中的任意一组就通过了
//              * /")*/
        
//           if(!empty($array)){
//               return true;
//           }
              
//         }else{
        
//             return false;
//         }
//         /*
//          * 虽然看起来复杂点,清楚理解!
//          * 如果有更好的,可以贴出来,分享快乐!
//          * */
        
//     }
// }