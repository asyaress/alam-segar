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
					<div class="card-action coin-tabs mb-3">
						<ul class="nav nav-tabs" role="tablist">
							<li class="nav-item">
								<a class="nav-link active" data-bs-toggle="tab" href="#AllEmployee">All Employee</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#ActiveEmployee">Active Employee</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#InactiveEmployee">Inactive Employee</a>
							</li>
						</ul>
					</div>
					<div class="d-flex align-items-center mb-3"> 
						<a href="javascript:void(0);" class="btn btn-secondary">+ New Employee</a>
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
									<div class="tab-pane active show" id="AllEmployee">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll">
															</div>
														</th>
														<th>Name</th>
														<th>Job Desk</th>
														<th>Schedule</th>
														<th>Contact</th>
														<th>Status</th>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/pic1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">James Sitepu</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Answering guest inquiries, directing phone calls, coordinating travel plans, and more.</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Louis Humbs</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Offer restaurant and activity recommendations and assist guests in arranging transportation</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Jackson Marquez</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Act as a liaison between guests and any department necessary including the kitchen, housekeeping, etc</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-danger">INACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/3.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Samantha William</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Anticipate guests needs in order to accommodate them and provide an exceptional guest experience</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/4.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">David Here</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Answering guest inquiries, directing phone calls, coordinating travel plans, and more.</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-danger">INACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/3.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Elina Moss</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Offer restaurant and activity recommendations and assist guests in arranging transportation</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
									<div class="tab-pane" id="ActiveEmployee">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all2">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll1">
															</div>
														</th>
														<th>Name</th>
														<th>Job Desk</th>
														<th>Schedule</th>
														<th>Contact</th>
														<th>Status</th>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/pic1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">James Sitepu</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Answering guest inquiries, directing phone calls, coordinating travel plans, and more.</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Louis Humbs</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Offer restaurant and activity recommendations and assist guests in arranging transportation</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Jackson Marquez</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Act as a liaison between guests and any department necessary including the kitchen, housekeeping, etc</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-danger">INACTIVE</span>
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
									<div class="tab-pane" id="InactiveEmployee">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all3">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll2">
															</div>
														</th>
														<th>Name</th>
														<th>Job Desk</th>
														<th>Schedule</th>
														<th>Contact</th>
														<th>Status</th>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/pic1.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">James Sitepu</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Answering guest inquiries, directing phone calls, coordinating travel plans, and more.</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-success">ACTIVE</span>
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
															<div class="concierge-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/avatar/2.jpg" alt="">
																<div>
																	<h5 class="fs-16 mb-0 text-nowrap"><a class="text-black" href="javascript:void(0);">Jackson Marquez</a></h5>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-14  text-nowrap">Joined on Aug 2th 2017</span>
																</div>
															</div>
														</td>
														<td class="job-desk3">
															<p class="mb-0">Act as a liaison between guests and any department necessary including the kitchen, housekeeping, etc</p>
														</td>
														<td>
															<div>
																<h5 class="text-nowrap">Monday, Friday</h5>
																<span class="tex-secondary text-nowrap">Check schedule</span>
															</div>
														</td>
														<td>
															<div>
																
																<span class="font-w600 text-nowrap"><i class="fas fa-phone-alt me-2 "></i>012 334 55512</span>
															</div>
														</td>
														<td>
															<span  class="font-w600 text-danger">INACTIVE</span>
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
				$('#reportrange span').php(start.format('D MMMM YYYY') + ' &nbsp - &nbsp ' + end.format('D MMMM YYYY'));
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