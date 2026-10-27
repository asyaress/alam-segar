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
				<div class="mt-4 d-flex justify-content-between align-items-center flex-wrap">
					<div class="card-action coin-tabs mb-2">
						<ul class="nav nav-tabs" role="tablist">
							<li class="nav-item">
								<a class="nav-link active" data-bs-toggle="tab" href="#AllRooms">All Rooms</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#ActiveEmployee">Active Employee</a>
							</li>
							<li class="nav-item">
								<a class="nav-link" data-bs-toggle="tab" href="#InactiveEmployee">Inactive Employee</a>
							</li>
						</ul>
					</div>
					<div class="d-flex align-items-center mb-2"> 
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
									<div class="tab-pane fade active show" id="AllRooms">
										<div class="table-responsive">
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all3">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll3">
															</div>
														</th>
														<th>Room Name</th>
														<th>Bed Type</th>
														<th>Room Floor</th>
														<th>Facilities</th>
														<th>Rate</th>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room4.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-success btn-md">ACTIVE</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room5.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room6.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div class="">
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-success btn-md">ACTIVE</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div class="">
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room6.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-success btn-md">ACTIVE</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div class="">
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all1">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll1">
															</div>
														</th>
														<th>Room Name</th>
														<th>Bed Type</th>
														<th>Room Floor</th>
														<th>Facilities</th>
														<th>Rate</th>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room6.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-success btn-md">ACTIVE</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
											<table class="table card-table display mb-4 shadow-hover table-responsive-lg" id="guestTable-all2">
												<thead>
													<tr>
														<th class="bg-none">
															<div class="form-check style-1">
															  <input class="form-check-input" type="checkbox" value="" id="checkAll2">
															</div>
														</th>
														<th>Room Name</th>
														<th>Bed Type</th>
														<th>Room Floor</th>
														<th>Facilities</th>
														<th>Rate</th>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room6.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-success btn-md">ACTIVE</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
															<div class="room-list-bx d-flex align-items-center">
																<img class="me-3 rounded" src="assets/images/room/room7.jpg" alt="">
																<div>
																	<span class=" text-secondary fs-14 d-block">#12341225</span>
																	<span class=" fs-16 font-w500 text-nowrap">Deluxe A-91234</span>
																</div>
															</div>
														</td>
														<td class="">
															<span class="fs-16 font-w500 text-nowrap">Double Bed</span>
														</td>
														<td>
															<div>
																
																<span class="fs-16 font-w500">Floor A-1</span>
															</div>
														</td>
														<td class="facility">
															<div>
																<span class="fs-16 comments">AC, Shower, Double Bed, Towel, Bathup, Coffee Set, LED TV, Wifi</span>
															</div>
														</td>
														<td>
															<div class="">
																<span class="mb-2">Price</span>	
																<span class="font-w500">$145<small class="fs-14 ms-2">/night</small></span>
															</div>
														</td>
														<td>
															<a href="javascript:void(0);" class="btn btn-danger btn-md">BOOKED</a>
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
</script>

</body>
</html>