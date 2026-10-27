<?php 
	 require_once __DIR__ . '/config/dz.php';
?>
<!DOCTYPE html>
<html lang="en">
<head>
   <!-- PAGE TITLE HERE -->
	<title><?php echo $DexignZoneSettings['site_level']['site_title'] ?></title>
	<?php include 'elements/meta.php';?>
	<!-- FAVICONS ICON -->
	<link rel="shortcut icon" type="image/png" href="<?php echo $DexignZoneSettings['site_level']['favicon']?>">
	<?php include 'elements/page-css.php'; ?>
</head>
<body>

	<!--*******************
		Preloader start
	********************-->
	 <?php include 'elements/pre-loader.php'; ?>
	<!--*******************
		Preloader end
	********************-->

	<!--**********************************
		Main wrapper start
	***********************************-->
	<div id="main-wrapper">

		<!--**********************************
			Nav header start
		***********************************-->
		<?php include 'elements/nav-header.php'; ?>
		<!--**********************************
			Nav header end
		***********************************-->
		
		<!--**********************************
			Chat box start
		***********************************-->
		<?php include 'elements/chatbox.php'; ?>
		<!--**********************************
			Chat box End
		***********************************-->
		
		<!--**********************************
			Header start
		***********************************-->
		<?php include 'elements/header.php'; ?>
		<!--**********************************
			Header end ti-comment-alt
		***********************************-->

		<!--**********************************
			Sidebar start
		***********************************-->
		<?php include 'elements/sidebar.php'; ?>
		<!--**********************************
			Sidebar end
		***********************************-->
		
		<!--**********************************
			Content body start
		***********************************-->
		<div class="content-body">
			<!-- row -->
			<div class="container-fluid">
				<div class="d-flex justify-content-between align-items-center flex-wrap">
					<div class="card-action coin-tabs mb-2">
						<ul class="nav nav-tabs" role="tablist">
							<li class="nav-item">
								<a class="nav-link active" data-bs-toggle="tab" href="#AllGuest">All Guest</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#Pending">Pending</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#Booked">Booked</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#Canceled">Canceled</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#Refund">Refund</a>
							</li>
						</ul>
					</div>
					<div class="d-flex align-items-center mb-2 flex-wrap"> 
						<div class="guest-calendar">
							<div id="reportrange" class="pull-right reportrange" style="width: 100%">
								<span></span><b class="caret"></b>
								<i class="fas fa-chevron-down ms-3"></i>
							</div>
						</div>
						<div class="newest ms-3">
							<select class="default-select">
								<option>Newest</option>
								<option>Oldest</option>
							</select>
						</div>	
					</div>
				</div>
				<div class="row mt-4">
					<div class="col-xl-12">
						<div class="card">
							<div class="card-body p-0">
								<div class="tab-content">	
									<div class="tab-pane active show" id="AllGuest">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover default-table table-responsive-lg" id="guestTable-all">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll">
															</div>
														</th>
														<th>Guest</th>
														<th>Order Date</th>
														<th>Check In</th>
														<th>Check Out</th>
														<th>Special Request</th>
														<th>Room Type</th>
														<th class="text-center">Status</th>
														<th class="bg-none"></th>
													</tr>
												</thead>
												<tbody>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Cahyadi Purnomo</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Cahyadi Purnomo</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Angela Smith</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/3.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Hendric Suratman</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/4.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Jajang Lawrence</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/5.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Kevin Stuart</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/6.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Steven</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/7.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Samantha</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
												</tbody>
											</table>
										</div>	
									</div>	
									<div class="tab-pane" id="Pending">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all1">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll1">
															</div>
														</th>
														<th>Guest</th>
														<th>Order Date</th>
														<th>Check In</th>
														<th>Check Out</th>
														<th>Special Request</th>
														<th>Room Type</th>
														<th class="text-center">Status</th>
														<th class="bg-none"></th>
													</tr>
												</thead>
												<tbody>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Cahyadi Purnomo</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Cahyadi Purnomo</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Angela Smith</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/3.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Hendric Suratman</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/4.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Jajang Lawrence</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
													
												</tbody>
											</table>
										</div>	
									</div>
									<div class="tab-pane" id="Booked">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all2">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll3">
															</div>
														</th>
														<th>Guest</th>
														<th>Order Date</th>
														<th>Check In</th>
														<th>Check Out</th>
														<th>Special Request</th>
														<th>Room Type</th>
														<th class="text-center">Status</th>
														<th class="bg-none"></th>
													</tr>
												</thead>
												<tbody>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Cahyadi Purnomo</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Cahyadi Purnomo</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>	
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Angela Smith</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
												</tbody>
											</table>
										</div>
									</div>
									<div class="tab-pane" id="Canceled">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all3">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll4">
															</div>
														</th>
														<th>Guest</th>
														<th>Order Date</th>
														<th>Check In</th>
														<th>Check Out</th>
														<th>Special Request</th>
														<th>Room Type</th>
														<th class="text-center">Status</th>
														<th class="bg-none"></th>
													</tr>
												</thead>
												<tbody>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Angela Smith</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
												</tbody>
											</table>
										</div>
									</div>
									<div class="tab-pane" id="Refund">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all4">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll5">
															</div>
														</th>
														<th>Guest</th>
														<th>Order Date</th>
														<th>Check In</th>
														<th>Check Out</th>
														<th>Special Request</th>
														<th>Room Type</th>
														<th class="text-center">Status</th>
														<th class="bg-none"></th>
													</tr>
												</thead>
												<tbody>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Angela Smith</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
													<tr>
														<td>
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="">
															</div>
														</td>
														<td>
															<div class="concierge-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Angela Smith</a></h5>
																	<span class="text-primary fs-14">#EMP-00025</span>
																</div>
															</div>
														</td>
														<td class="text-nowrap">
															<span>Oct 30th 2020 09:21 AM</span>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 2th, 2020</h5>
																<span>9.46 PM</span>
															</div>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Nov 4th, 2020</h5>
																<span>6.12 PM</span>
															</div>
														</td>
														<td class="request">
															<a href="javascript:void(0);" class="btn  btn-sm">View Notes</a>
														</td>
														<td>
															<span class="font-w500">Deluxe A - 02</span>
														</td>
														<td>
															<div class="request">
																<a href="javascript:void(0);" class="btn btn-md text-primary">Refund</a>
															</div>
														</td>
														<td>
															<div class="dropdown dropend">
																<a href="javascript:void(0);" class="btn-link" data-bs-toggle="dropdown" aria-expanded="false">
																	<svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
																		<path d="M11 12C11 12.5523 11.4477 13 12 13C12.5523 13 13 12.5523 13 12C13 11.4477 12.5523 11 12 11C11.4477 11 11 11.4477 11 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M18 12C18 12.5523 18.4477 13 19 13C19.5523 13 20 12.5523 20 12C20 11.4477 19.5523 11 19 11C18.4477 11 18 11.4477 18 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																		<path d="M4 12C4 12.5523 4.44772 13 5 13C5.55228 13 6 12.5523 6 12C6 11.4477 5.55228 11 5 11C4.44772 11 4 11.4477 4 12Z" stroke="#262626" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
																	</svg>
																</a>
																<div class="dropdown-menu">
																	<a class="dropdown-item" href="javascript:void(0);">Edit</a>
																	<a class="dropdown-item" href="javascript:void(0);">Delete</a>
																</div>
															</div>
														</td>
													</tr>
												</tbody>
											</table>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
		<!--**********************************
			Content body end
		***********************************-->
		
		
		
		<!--**********************************
			Footer start
		***********************************-->
		<?php include 'elements/footer.php'; ?>
		<!--**********************************
			Footer end
		***********************************-->

		<!--**********************************
		   Support ticket button start
		***********************************-->
		
		<!--**********************************
		   Support ticket button end
		***********************************-->


	</div>
	<!--**********************************
		Main wrapper end
	***********************************-->

	<!--**********************************
		Scripts
	***********************************-->
	<!-- Required vendors -->
	<?php include 'elements/page-js.php'; ?>
	
	<script>
		$(function() {

			var start = moment().subtract(29, 'days');
			var end = moment();

			function cb(start, end) {
				$('#reportrange span').html(start.format('D MMMM YYYY') + ' &nbsp - &nbsp ' + end.format('D MMMM YYYY'));
    }

    $('#reportrange').daterangepicker({
        startDate: start,
        endDate: end,
        ranges: {
           'Today': [moment(), moment()],
           'Yesterday': [moment().subtract(1, 'days'), moment().subtract(1, 'days')],
           'Last 7 Days': [moment().subtract(6, 'days'), moment()],
           'Last 30 Days': [moment().subtract(29, 'days'), moment()],
           'This Month': [moment().startOf('month'), moment().endOf('month')],
           'Last Month': [moment().subtract(1, 'month').startOf('month'), moment().subtract(1, 'month').endOf('month')]
        }
    }, cb);

    cb(start, end);
    
});
</script>z

</body>
</html>