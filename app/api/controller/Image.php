<?php    
namespace app\api\controller;
class Image{
    
    public function getImage(){
        
// 图片一  
    $path1 = '../public/static/admin/images/1.jpg';  
    // 图片二  
    $path2 = '../public/static/admin/images/3.jpg';  
    // 创建图片对象  
    $image1 = imagecreatefromjpeg($path1);  
    $image2 = imagecreatefromjpeg($path2);  
    // 合成图片  
    imagecopymerge($image1, $image2, 0, 0, 0, 0, imagesx($image_2)/2, imagesy($image_2), 100);  
    // 输出合成图片  
    $imagename = rand(1000,9999).'-'.time().'jpg';
    header('Content-Type:image/jpeg');
    imagejpeg($image1,'../public/static/admin/images/a.jpg');
    
    //释放内存
    imagedestroy($image1);
    imagedestroy($image2);
    }
}