
<?php 
	 require_once __DIR__ . '/config/dz.php';
?>
<!DOCTYPE html>
<html lang="en" class="h-100">

<head>
   <!-- PAGE TITLE HERE -->
	<title><?php echo $DexignZoneSettings['site_level']['site_title'] ?></title>
	<?php include 'elements/meta.php';?>
	<!-- FAVICONS ICON -->
	<link rel="shortcut icon" type="image/png" href="<?php echo $DexignZoneSettings['site_level']['favicon']?>">
	<?php include 'elements/page-css.php'; ?>
</head>

<body class="vh-100">    
    <div class="authincation h-100">
        <div class="container h-100">
            <div class="row justify-content-center h-100 align-items-center">
                <div class="col-md-5">
                    <div class="form-input-content text-center error-page">
                        <h1 class="error-text fw-bold">500</h1>
                        <h4><i class="fa fa-times-circle text-danger"></i> Internal Server Error</h4>
                        <p>You do not have permission to view this resource</p> 
						<div>
                            <a class="btn btn-primary" href="./index.php">Back to Home</a>
                        </div>	
                    </div>
                </div>
            </div>
        </div>
    </div>

<!--**********************************
	Scripts
***********************************-->
<!-- Required vendors -->
<script src="./vendor/global/global.min.js"></script>
<script src="./js/custom.min.js"></script>
<script src="./js/dlabnav-init.js"></script>
<script src="./js/styleSwitcher.js"></script>
</body>
</html>