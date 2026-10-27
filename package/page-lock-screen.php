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
                <div class="col-md-6">
                    <div class="authincation-content">
                        <div class="row no-gutters">
                            <div class="col-xl-12">
                                <div class="auth-form">
									<div class="text-center mb-3">
										<a href="index.php"><img src="assets/images/logo-full.png" alt=""></a>
									</div>
                                    <h4 class="text-center mb-4">Account Locked</h4>
                                    <form action="index.php">
                                        <div class="mb-3">
                                            <label><strong>Password</strong></label>
                                            <input type="password" class="form-control" value="Password">
                                        </div>
                                        <div class="text-center">
                                            <button type="submit" class="btn btn-primary btn-block">Unlock</button>
                                        </div>
                                    </form>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <!-- #/ container -->
    <!-- Common JS -->
    <script src="./vendor/global/global.min.js"></script>
    <!-- Custom script -->
    <script src="./vendor/dlabnav/dlabnav.min.js"></script>
    <script src="./js/custom.min.js"></script>
    <script src="./js/dlabnav-init.js"></script>
	<script src="./js/styleSwitcher.js"></script>
</body>
</html>