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
				<div class="row mt-4">
					<div class="col-xl-12">
						<div class="row">
							<div class="col-xl-12">
								<div>
									<div class="reviews-slider owl-carousel owl-carousel owl-loaded owl-drag">
										<div class="items">
											<div class="customers review-slider">
												<p class="fs-16">Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam</p>
												<div class="d-flex justify-content-between align-items-center mt-5">
													<div class="customer-profile d-flex ">
													<img src="assets/images/users/user5.jpg" alt="">
													<div class="ms-3">
														<h5 class="mb-0"><a href="javascript:void(0);">Kusnaidi Anderson</a></h5>
														<span>4m ago</span>
													</div>
													</div>
													<div class="customer-button text-nowrap">
														<a href="javascript:void(0);"><i class="far fa-check-circle text-success"></i></a>
														<a href="javascript:void(0);"><i class="far fa-times-circle text-danger"></i></a>
													</div>
												</div>
											</div>
										</div>
										<div class="items">
											<div class="customers review-slider">
												<p class="fs-16">Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam</p>
												<div class="d-flex justify-content-between align-items-center mt-5">
													<div class="customer-profile d-flex ">
													<img src="assets/images/users/user6.jpg" alt="">
													<div class="ms-3">
														<h5 class="mb-0"><a href="javascript:void(0);">Kusnaidi Anderson</a></h5>
														<span>4m ago</span>
													</div>
													</div>
													<div class="customer-button text-nowrap">
														<a href="javascript:void(0);"><i class="far fa-check-circle text-success"></i></a>
														<a href="javascript:void(0);"><i class="far fa-times-circle text-danger"></i></a>
													</div>
												</div>
											</div>
										</div>
										<div class="items">
											<div class="customers review-slider">
												<p class="fs-16">Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam</p>
												<div class="d-flex justify-content-between align-items-center mt-5">
													<div class="customer-profile d-flex ">
													<img src="assets/images/users/user7.jpg" alt="">
													<div class="ms-3">
														<h5 class="mb-0"><a href="javascript:void(0);">Kusnaidi Anderson</a></h5>
														<span>4m ago</span>
													</div>
													</div>
													<div class="customer-button text-nowrap">
														<a href="javascript:void(0);"><i class="far fa-check-circle text-success"></i></a>
														<a href="javascript:void(0);"><i class="far fa-times-circle text-danger"></i></a>
													</div>
												</div>
											</div>
										</div>
										<div class="items">
											<div class="customers review-slider">
												<p class="fs-16">Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam</p>
												<div class="d-flex justify-content-between align-items-center mt-5">
													<div class="customer-profile d-flex ">
													<img src="assets/images/users/user5.jpg" alt="">
													<div class="ms-3">
														<h5 class="mb-0"><a href="javascript:void(0);">Kusnaidi Anderson</a></h5>
														<span>4m ago</span>
													</div>
													</div>
													<div class="customer-button text-nowrap">
														<a href="javascript:void(0);"><i class="far fa-check-circle text-success"></i></a>
														<a href="javascript:void(0);"><i class="far fa-times-circle text-danger"></i></a>
													</div>
												</div>
											</div>
										</div>
									</div>
								</div>
							</div>
							<div class="my-4 d-flex justify-content-between align-items-center flex-wrap">
								<div class="card-action coin-tabs mb-4">
									<ul class="nav nav-tabs" role="tablist">
										<li class="nav-item">
											<a class="nav-link active" data-bs-toggle="tab" href="#AllCustomerReviews">All Customer Reviews</a>
										</li>
										<li class="nav-item">
											<a class="nav-link" data-bs-toggle="tab" href="#Published">Published</a>
										</li>
										<li class="nav-item">
											<a class="nav-link" data-bs-toggle="tab" href="#InactiveEmployee">Archived</a>
										</li>
									</ul>
								</div>
								<div class="newest mb-4">
									<select class="default-select">
										<option>Newest</option>
										<option>Oldest</option>
									</select>
								</div>
							</div>
							<div class="col-xl-12">
								<div class="card">
									<div class="card-body p-0">
										<div class="tab-content">	
											<div class="tab-pane active show" id="AllCustomerReviews">
												<div class="table-responsive">
													<table class="table card-table display mb-4 shadow-hover table-responsive-lg review-tbl" id="guestTable-all">
														<thead>
															<tr>
																<th class="bg-none">
																	<div class="form-check style-1">
																	  <input class="form-check-input" type="checkbox" value="" id="checkAll">
																	</div>
																</th>
																<th>Order ID</th>
																<th>Date</th>
																<th>Customer</th>
																<th class="text-center">Comment</th>
																<th class="text-center">Action</th>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
																	</div>
																</td>
															</tr>	
														</tbody>
													</table>	
												</div>	
											</div>	
											<div class="tab-pane" id="Published">
												<div class="table-responsive">
													<table class="table card-table display mb-4 shadow-hover table-responsive-lg review-tbl" id="guestTable-all1">
														<thead>
															<tr>
																<th class="bg-none">
																	<div class="form-check style-1">
																	  <input class="form-check-input" type="checkbox" value="" id="checkAll">
																	</div>
																</th>
																<th>Order ID</th>
																<th>Date</th>
																<th>Customer</th>
																<th class="text-center">Comment</th>
																<th class="text-center">Action</th>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
																	</div>
																</td>
															</tr>	
														</tbody>	
													</table>	
												</div>
											</div>
											<div class="tab-pane" id="InactiveEmployee">
												<div class="table-responsive">
													<table class="table card-table display mb-4 shadow-hover table-responsive-lg review-tbl" id="guestTable-all2">
														<thead>
															<tr>
																<th class="bg-none">
																	<div class="form-check style-1">
																	  <input class="form-check-input" type="checkbox" value="" id="checkAll">
																	</div>
																</th>
																<th>Order ID</th>
																<th>Date</th>
																<th>Customer</th>
																<th class="text-center">Comment</th>
																<th class="text-center">Action</th>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
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
																	<span>#000123456</span>
																</td>
																<td>
																	<span class="text-nowrap">Nov 21th 2020 09:21 AM</span>
																</td>
																<td>
																	<span class="text-nowrap">James Sitepu</span>
																</td>
																<td class="job-desk1">
																	<span>
																		<ul class="stars">
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																			<li><a href="javascript:void(0);"><i class="fas fa-star text-secondary"></i></a></li>
																		</ul>
																	</span>
																	<span>
																		<p class="fs-16">
																			We recently had dinner with friends at Dimas Can Zheng and we all walked away with a great experience. Good food, pleasant environment, personal attention through all the evening. Thanks to the team and we will be back
																		</p>
																	</span>
																</td>
																<td>
																	<div class="d-flex">
																		<span   class="text-success fs-14 font-w500 me-3">Publish</span>
																		<span   class="text-danger fs-14 font-w500" >Archive</span>
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
		function TravlCarousel()
			{

				/*  testimonial one function by = owl.carousel.js */
				jQuery('.reviews-slider').owlCarousel({
					loop:false,
					margin:15,
					nav:false,
					autoplaySpeed: 3000,
					navSpeed: 3000,
					paginationSpeed: 3000,
					slideSpeed: 3000,
					smartSpeed: 3000,
					autoplay: false,
					animateOut: 'fadeOut',
					dots:false,
					navigation:false,
					navText: ['', ''],
					responsive:{
						0:{
							items:1
						},
						
						768:{
							items:2
						},			
						
						1400:{
							items:2
						},
						1600:{
							items:3
						},
						1750:{
							items:3
						}
					}
				})
			}

			jQuery(window).on('load',function(){
				setTimeout(function(){
					TravlCarousel();
				}, 1000); 
			});
	</script>

</body>
</html>