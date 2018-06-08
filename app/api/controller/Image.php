<?php    
namespace app\api\controller;
use think\Controller;
class Image{
    
    public function imageMerge($arr){
        $arr = input('post.');
        dump($name);
        die();
        // 图片一  
        $path1 = '../public/static/admin/images/3.jpg';  
        // 图片二   
        $path2 = '../public/static/admin/images/1.jpg';  
        // 创建图片对象  1
        $image1 = imagecreatefromjpeg($path1);   
        $image2 = imagecreatefromjpeg($path2); 
        // 合成图片  
        imagecopymerge($image1, $image2, 0, 0, 0, 0, imagesx($image2), imagesy($image2), 100);  
        // 输出合成图片  
        // $imagename = rand(1000,9999).'-'.time().'jpg';
      
        exit();   
        $font = '../public/static/admin/lib/fangzheng.ttf';
        $black = imagecolorallocate($image1, 100, 100,100);//字体颜色 RGB
        $fontSize = 20;
        $circleSize = 0;  
        $fontBox = imagettfbbox($fontSize, 0, $font, $name);
        imagefttext($image1,$fontSize,$circleSize,50,50,$black,$font,$arr['name']);
        
        
        
        header('Content-Type:image/jpeg'); 
        imagejpeg($image1,'../public/static/admin/images/a.jpg');
        
        //释放内存
        imagedestroy($image1);
        imagedestroy($image2);
        
        return "http://kxs.ruohua.club/static/admin/images/a.jpg";
    
    }
//     public function textMerge(){
//         $name = input('post.name');
//         $img = rand(1,5).'.jpg';
//         $path = "..public/static/admin/images/$img";
//         $image = imagecreatefromjpeg($path);
//         dump($image);
//         $black = imagecolorallocate($img, 100, 100,100);//字体颜色 RGB
//         $ttf = '../public/static/admin/lib/fangzheng.ttf';
//         imagefttext($image,16,355,20,20,$black,$ttf,$name);
//         $idimg = rand(1000,9999).'jpg';

//         header('Content-Type:image/jpeg');
//         imagejpeg($image,"../public/static/admin/images/$idimg");
        
//         //释放内存
//         imagedestroy($image); 
        
//         return "http://kxs.ruohua.club/static/admin/images/$idimg";
        
        
//     }
}